import hashlib
import json
import subprocess
import tomllib
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

import boto3
import httpx
import rich
import typer
from rich.progress import BarColumn, Progress, SpinnerColumn, TextColumn

LEGAL_SETS = [
    "4ed",
    "ice",
    "chr",
    "hml",
    "all",
    "mir",
    "vis",
    "5ed",
    "wth",
    "tmp",
    "sth",
    "exo",
    "usg",
    "ulg",
    "6ed",
    "uds",
    "mmq",
    "nem",
    "pcy",
    "inv",
    "pls",
    "7ed",
    "apc",
    "ody",
    "tor",
    "jud",
    "ons",
    "lgn",
    "scg",
]


@dataclass
class AWSConfig:
    aws_access_key: str
    aws_secret_access_key: str
    s3_bucket_name: str
    s3_bucket_prefix: str
    s3_bucket_region: str


app = typer.Typer()


def _load_config() -> AWSConfig:
    config_path = Path(__file__).parent / "config.toml"
    with open(config_path, "rb") as f:
        config = tomllib.load(f)
    return AWSConfig(**config["aws"])


def _get_s3_client():
    config = _load_config()
    return boto3.client(
        "s3",
        aws_access_key_id=config.aws_access_key,
        aws_secret_access_key=config.aws_secret_access_key,
        region_name=config.s3_bucket_region,
    )


def _calculate_file_etag(file_path: Path) -> str:
    md5 = hashlib.md5()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            md5.update(chunk)
    return md5.hexdigest()


def _get_todays_bulk_file() -> dict:
    current_dir = Path(__file__).parent
    today = datetime.now().strftime("%Y-%m-%d")
    todays_file = current_dir / f"default_cards_{today}.json"

    if not todays_file.exists():
        url = "https://api.scryfall.com/bulk-data"
        response = httpx.get(url)
        data = response.json()

        bulk_url = ""
        for o in data["data"]:
            if o["type"] == "default_cards":
                bulk_url = o["download_uri"]
                break

        if not bulk_url:
            raise ValueError("Could not find default_cards bulk data URL.")

        response = httpx.get(bulk_url)
        with open(todays_file, "w") as f:
            f.write(response.text)

    with open(todays_file, "r") as f:
        return json.load(f)


@app.command()
def event(date: str, matches: int, players: list[str]) -> None:
    event_data = {
        "name": f"Onsdagstävling {date}",
        "date": date,
        "rounds": 4,
        "player_info": {},
        "matches": [],
    }
    for player in players:
        event_data["player_info"][player] = {
            "deck": "",
            "decklist": "",
        }

    for match in range(1, matches + 1):
        event_data["matches"].append(
            {
                "player_1": "",
                "player_2": "",
                "result": "",
            }
        )

    event_path = Path(__file__).parent.parent / "input" / "events" / f"{date}.json"

    with open(event_path, "w") as f:
        json.dump(event_data, f, indent=4, ensure_ascii=False)


@app.command()
def rename(old_name: str, new_name: str) -> None:
    """
    Rename a player in all event files from old_name to new_name.
    """
    events_dir = Path(__file__).parent.parent / "input" / "events"
    event_files = events_dir.glob("*.json")

    for event_file in event_files:
        with open(event_file, "r") as f:
            event_data = json.load(f)

        if old_name in event_data["player_info"]:
            event_data["player_info"][new_name] = event_data["player_info"].pop(old_name)

            for match in event_data["matches"]:
                if match["player_1"] == old_name:
                    match["player_1"] = new_name
                if match["player_2"] == old_name:
                    match["player_2"] = new_name

            with open(event_file, "w") as f:
                json.dump(event_data, f, indent=4, ensure_ascii=False)

            rich.print(f"Renamed {old_name} to {new_name} in event file {event_file.name}")


@app.command()
def db() -> None:
    """
    Generate a card database (db.json) from the provided scryfall default cards file.
    """
    cards = _get_todays_bulk_file()

    cards_db = []
    unique_cards = set()

    white_border_cards = {}

    for card in cards:
        name = card["name"]
        set_code = card.get("set", "")
        legality = card.get("legalities", {}).get("premodern", "")
        border = card.get("border_color")

        finishes = card.get("finishes", [])
        if "nonfoil" not in finishes:
            continue

        if name in unique_cards:
            continue

        if legality == "not_legal":
            continue

        if set_code not in LEGAL_SETS:
            continue

        card_type = "other"
        if "Land" in card.get("type_line", ""):
            card_type = "land"
        elif "Creature" in card.get("type_line", ""):
            card_type = "creature"

        db_card = {
            "name": name,
            "image_url": card.get("image_uris", {}).get("border_crop", ""),
            "legality": legality,
            "card_type": card_type,
        }

        if border == "white":
            if name in white_border_cards:
                continue

            white_border_cards[name] = db_card
            continue

        cards_db.append(db_card)
        unique_cards.add(name)

    for name, db_card in white_border_cards.items():
        if name in unique_cards:
            continue

        cards_db.append(db_card)
        unique_cards.add(name)

    with open(Path(__file__).parent.parent / "files" / "db.json", "w") as f:
        json.dump(sorted(cards_db, key=lambda x: x["name"]), f, indent=2)

    card_count = len(cards_db)
    rich.print(f"Card database updated with {card_count} cards.")
    if card_count != 5408:
        rich.print("Warning: The expected card count is 5408. Please verify the database.")


@app.command()
def decklist(date: str, player: str, deck: str = "") -> None:
    """
    Update the decklist ID for a player in a specific event file.
    """
    player_slug = player.lower().replace(" ", "-")
    decklist_path = Path(__file__).parent.parent / "input" / "decklists" / f"{date}-{player_slug}.txt"
    if decklist_path.exists():
        rich.print(f"Decklist file {decklist_path} already exists.")
        return

    clipboard_data = subprocess.run(["pbpaste"], capture_output=True, text=True).stdout

    with open(decklist_path, "w") as f:
        f.write(clipboard_data)

    event_path = Path(__file__).parent.parent / "input" / "events" / f"{date}.json"

    with open(event_path, "r") as f:
        event_data = json.load(f)

    if player in event_data["player_info"]:
        event_data["player_info"][player]["decklist"] = f"{date}-{player_slug}"
        if deck:
            event_data["player_info"][player]["deck"] = deck

        with open(event_path, "w") as f:
            json.dump(event_data, f, indent=4, ensure_ascii=False)

        rich.print(f"Updated decklist for {player} in event file {event_path.name}")
    else:
        rich.print(f"Player {player} not found in event file {event_path.name}")


@app.command()
def download() -> None:
    """
    Download files from S3 bucket, only downloading files with changed ETags.
    """
    config = _load_config()
    s3_client = _get_s3_client()

    bucket_name = config.s3_bucket_name
    bucket_prefix = config.s3_bucket_prefix

    etag_file = Path(__file__).parent / "etags.json"
    if etag_file.exists():
        with open(etag_file, "r") as f:
            stored_etags = json.load(f)
    else:
        stored_etags = {}

    paginator = s3_client.get_paginator("list_objects_v2")
    pages = paginator.paginate(Bucket=bucket_name, Prefix=bucket_prefix)

    files_to_download = []
    current_etags = {}

    for page in pages:
        if "Contents" not in page:
            continue

        for obj in page["Contents"]:
            key = obj["Key"]
            etag = obj["ETag"].strip('"')

            if key.endswith("/"):
                continue

            relative_path = key.replace(f"{bucket_prefix}/", "", 1)
            current_etags[relative_path] = etag

            stored_etag = stored_etags.get(relative_path)
            if stored_etag != etag:
                files_to_download.append((key, etag))

    if not files_to_download:
        typer.echo("All files are up to date.")
        return

    project_root = Path(__file__).parent.parent

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
    ) as progress:
        task = progress.add_task(f"Downloading {len(files_to_download)} file(s)", total=len(files_to_download))

        for key, etag in files_to_download:
            relative_path = key.replace(f"{bucket_prefix}/", "", 1)
            local_path = project_root / "input" / relative_path

            local_path.parent.mkdir(parents=True, exist_ok=True)

            s3_client.download_file(bucket_name, key, str(local_path))
            progress.update(task, advance=1, description=f"Downloaded: {relative_path}")

    with open(etag_file, "w") as f:
        json.dump(current_etags, f, indent=2)

    typer.echo(f"Successfully downloaded {len(files_to_download)} file(s).")


def _updated_input_files() -> list[Path]:
    """
    Check local file ETags in input directory against stored ETags.
    Returns a list of Paths for files that do not match etags.
    """
    etag_file = Path(__file__).parent / "etags.json"
    if etag_file.exists():
        with open(etag_file, "r") as f:
            stored_etags = json.load(f)
    else:
        stored_etags = {}

    project_root = Path(__file__).parent.parent
    input_dir = project_root / "input"

    mismatched_files = []

    for file_path in input_dir.rglob("*"):
        if file_path.is_file() and not file_path.name.startswith("."):
            relative_path = file_path.relative_to(input_dir)
            local_etag = _calculate_file_etag(file_path)
            stored_etag = stored_etags.get(str(relative_path))

            if stored_etag != local_etag:
                mismatched_files.append(file_path)

    return mismatched_files


@app.command()
def upload() -> None:
    """
    Upload files from input directory to S3 bucket, only uploading files with changed ETags.
    """
    config = _load_config()
    s3_client = _get_s3_client()

    bucket_name = config.s3_bucket_name
    bucket_prefix = config.s3_bucket_prefix

    files_to_upload = _updated_input_files()

    if not files_to_upload:
        typer.echo("All files are up to date.")
        return

    project_root = Path(__file__).parent.parent
    input_dir = project_root / "input"

    etag_file = Path(__file__).parent / "etags.json"
    if etag_file.exists():
        with open(etag_file, "r") as f:
            stored_etags = json.load(f)
    else:
        stored_etags = {}

    uploaded_count = 0

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
    ) as progress:
        task = progress.add_task(f"Uploading {len(files_to_upload)} file(s)", total=len(files_to_upload))

        for file_path in files_to_upload:
            relative_path = file_path.relative_to(input_dir)
            s3_key = f"{bucket_prefix}/{relative_path}"

            try:
                s3_client.upload_file(str(file_path), bucket_name, s3_key)

                local_etag = _calculate_file_etag(file_path)
                stored_etags[str(relative_path)] = local_etag
                uploaded_count += 1

                progress.update(task, advance=1, description=f"Uploaded: {relative_path}")
            except Exception as e:
                progress.stop()
                error_msg = str(e)
                if "AccessDenied" in error_msg or "403" in error_msg:
                    rich.print("\n[red]Error: Access denied when uploading to S3 bucket.[/red]")
                    rich.print(f"[yellow]You may not have upload permissions for bucket '{bucket_name}'.[/yellow]")
                else:
                    rich.print(f"\n[red]Error uploading {relative_path}: {error_msg}[/red]")
                rich.print(
                    f"[yellow]Aborted upload. {uploaded_count} of {len(files_to_upload)} file(s) were uploaded successfully.[/yellow]"
                )
                return

    with open(etag_file, "w") as f:
        json.dump(stored_etags, f, indent=2)

    typer.echo(f"Successfully uploaded {uploaded_count} file(s).")


if __name__ == "__main__":
    app()
