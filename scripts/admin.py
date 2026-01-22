import json
import subprocess
from pathlib import Path

import rich
import typer

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

app = typer.Typer()


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
def db(filename: Path) -> None:
    """
    Generate a card database (db.json) from the provided scryfall default cards file.
    """
    current_dir = Path(__file__).parent
    with open(filename, "r") as f:
        cards = json.load(f)

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

    with open(current_dir.parent / "files" / "db.json", "w") as f:
        json.dump(sorted(cards_db, key=lambda x: x["name"]), f, indent=2)

    card_count = len(cards_db)
    rich.print(f"Card database updated with {card_count} cards.")
    if card_count != 5408:
        rich.print("Warning: The expected card count is 5408. Please verify the database.")


@app.command()
def decklist(date: str, player: str) -> None:
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

        with open(event_path, "w") as f:
            json.dump(event_data, f, indent=4, ensure_ascii=False)

        rich.print(f"Updated decklist for {player} in event file {event_path.name}")
    else:
        rich.print(f"Player {player} not found in event file {event_path.name}")


if __name__ == "__main__":
    app()
