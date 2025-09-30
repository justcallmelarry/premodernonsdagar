import json
from pathlib import Path

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


IGNORE_COMBO = {}


def main() -> None:
    """
    Generate a card database (db.json) from the provided scryfall default cards file.
    """
    current_dir = Path(__file__).parent
    with open(current_dir / "default-cards.json", "r") as f:
        cards = json.load(f)

    cards_db = []
    unique_cards = set()

    for card in cards:
        name = card["name"]
        set_code = card.get("set", "")
        legality = card.get("legalities", {}).get("premodern", "")

        if name in unique_cards:
            continue

        if legality == "not_legal":
            continue

        if set_code not in LEGAL_SETS:
            continue

        cards_db.append(
            {
                "name": name,
                "image_url": card.get("image_uris", {}).get("normal", ""),
                "legality": card.get("legalities", {}).get("premodern"),
            }
        )
        unique_cards.add(name)

    # Save to db.json
    with open(current_dir / "db.json", "w") as f:
        json.dump(cards_db, f, indent=2)

    print(f"Card database updated with {len(cards_db)} cards. (should be 5408)")


if __name__ == "__main__":
    main()
