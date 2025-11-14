# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "typer",
# ]
# ///
import json
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


def db() -> None:
    """
    Generate a card database (db.json) from the provided scryfall default cards file.
    """
    current_dir = Path(__file__).parent
    with open(current_dir / "default-cards.json", "r") as f:
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
        rich.print(
            "Warning: The expected card count is 5408. Please verify the database."
        )


if __name__ == "__main__":
    app()
