# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "typer",
# ]
# ///
import json
from pathlib import Path

import typer


def main(date: str, matches: int, players: list[str]) -> None:
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


if __name__ == "__main__":
    typer.run(main)
