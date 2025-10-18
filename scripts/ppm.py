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


def main() -> None:
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

        if name in ("Plains", "Island", "Swamp", "Mountain", "Forest"):
            continue

        price_trend = card.get("prices", {}).get("eur", None)
        if price_trend is not None:
            try:
                price_trend = float(price_trend)
            except ValueError:
                price_trend = None

        db_card = {
            "name": name,
            "image_url": card.get("image_uris", {}).get("normal", ""),
            "price_trend": price_trend,
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

    with open(current_dir / "db.json", "w") as f:
        json.dump(sorted(cards_db, key=lambda x: x["name"]), f, indent=2)

    card_count = len(cards_db)
    print(f"Card database updated with {card_count} cards.")
    if card_count != 5403:
        print("Warning: The expected card count is 5403. Please verify the database.")


if __name__ == "__main__":
    main()
