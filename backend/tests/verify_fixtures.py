"""Fast, dependency-free validation for shared test and mock data."""
from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DATA = ROOT / "data"


def load(path: Path):
    with path.open(encoding="utf-8") as stream:
        return json.load(stream)


def assert_envelope(response: dict, name: str) -> None:
    assert set(response) == {"success", "data", "error"}, f"{name}: invalid envelope keys"
    if response["success"]:
        assert response["error"] is None, f"{name}: success response has error"
        assert response["data"] is not None, f"{name}: success response lacks data"
    else:
        assert response["data"] is None, f"{name}: error response has data"
        assert isinstance(response["error"], dict) and {"code", "message"} <= set(response["error"]), f"{name}: invalid error"


def main() -> None:
    final_sprint = load(DATA / "fixtures" / "final_sprint.json")
    assert final_sprint["dataset"]["dataset_id"].startswith("ds_")
    assert final_sprint["dataset"]["file_count"] >= len(final_sprint["files"])
    for question in final_sprint["questions"]:
        assert question["dataset_id"] == final_sprint["dataset"]["dataset_id"]
        assert question["question_id"].startswith("q_")
        assert question["difficulty"] in {"easy", "medium", "hard"}

    homework = load(DATA / "fixtures" / "homework.json")
    assert homework["question"]["homework_id"] == homework["homework"]["homework_id"]
    assert homework["hint_request"]["help_level"] == "direction"

    settings = load(DATA / "fixtures" / "settings.json")
    assert settings["response_style"] in {"concise", "normal", "detailed"}
    assert settings["answer_policy"] in {"hint_first", "balanced", "direct_answer"}

    for name, response in load(DATA / "mock" / "responses.json").items():
        assert_envelope(response, name)

    for path in DATA.rglob("*.json"):
        load(path)
    print("Fixture validation passed.")


if __name__ == "__main__":
    main()
