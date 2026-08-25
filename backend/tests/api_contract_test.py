"""Black-box API contract checks; uses only the Python standard library."""
from __future__ import annotations

import json
import os
import sys
import unittest
from datetime import date, timedelta
from urllib.error import HTTPError
from urllib.request import Request, urlopen

BASE_URL = os.getenv("API_BASE_URL", "http://localhost:8080").rstrip("/")


def request(method: str, path: str, payload: dict | None = None) -> tuple[int, dict]:
    body = json.dumps(payload).encode() if payload is not None else None
    headers = {"Content-Type": "application/json"} if body else {}
    req = Request(f"{BASE_URL}{path}", data=body, headers=headers, method=method)
    try:
        with urlopen(req, timeout=10) as response:
            return response.status, json.loads(response.read())
    except HTTPError as error:
        return error.code, json.loads(error.read())


class APIContractTests(unittest.TestCase):
    dataset_id: str

    def assert_envelope(self, payload: dict, success: bool) -> None:
        self.assertEqual(set(payload), {"success", "data", "error"})
        self.assertIs(payload["success"], success)
        if success:
            self.assertIsNone(payload["error"])
            self.assertIsNotNone(payload["data"])
        else:
            self.assertIsNone(payload["data"])
            self.assertIsInstance(payload["error"], dict)
            self.assertTrue({"code", "message"} <= set(payload["error"]))

    def test_01_create_dataset(self) -> None:
        _, payload = request("POST", "/api/v1/final-sprint/datasets", {"name": "契约测试资料集", "course": "高等数学"})
        self.assert_envelope(payload, True)
        self.assertRegex(payload["data"]["dataset_id"], r"^ds_")
        self.assertEqual(payload["data"]["status"], "created")
        type(self).dataset_id = payload["data"]["dataset_id"]

    def test_02_reject_invalid_requests(self) -> None:
        cases = [
            ("POST", "/api/v1/final-sprint/datasets", {}),
            ("POST", "/api/v1/final-sprint/datasets", {"name": "", "course": "高等数学"}),
            ("POST", "/api/v1/chat", {"message": ""}),
            ("PUT", "/api/v1/settings/agent", {"response_style": "verbose", "personality": "encouraging", "answer_policy": "answer_everything"}),
        ]
        for method, path, body in cases:
            with self.subTest(path=path, body=body):
                _, payload = request(method, path, body)
                self.assert_envelope(payload, False)
                self.assertEqual(payload["error"]["code"], "INVALID_REQUEST")

    def test_03_not_found_resources(self) -> None:
        for path, expected in [("/api/v1/tasks/task_missing", "TASK_NOT_FOUND"), ("/api/v1/final-sprint/datasets/ds_missing/analysis", "DATASET_NOT_FOUND")]:
            with self.subTest(path=path):
                _, payload = request("GET", path)
                self.assert_envelope(payload, False)
                self.assertEqual(payload["error"]["code"], expected)

    def test_04_plan_boundaries(self) -> None:
        dataset_id = self.dataset_id
        tomorrow = (date.today() + timedelta(days=1)).isoformat()
        valid = {"exam_date": tomorrow, "daily_study_hours": 0.5, "current_level": "medium"}
        _, payload = request("POST", f"/api/v1/final-sprint/datasets/{dataset_id}/plan", valid)
        self.assert_envelope(payload, True)
        for body in ({"exam_date": "not-a-date", "daily_study_hours": 4}, {"exam_date": "2020-01-01", "daily_study_hours": 4}, {"exam_date": tomorrow, "daily_study_hours": 0}):
            with self.subTest(body=body):
                _, payload = request("POST", f"/api/v1/final-sprint/datasets/{dataset_id}/plan", body)
                self.assert_envelope(payload, False)
                self.assertEqual(payload["error"]["code"], "INVALID_REQUEST")

    def test_05_chat_and_settings_contract(self) -> None:
        _, payload = request("POST", "/api/v1/chat", {"message": "解释贝叶斯公式", "course": "概率论"})
        self.assert_envelope(payload, True)
        self.assertIsInstance(payload["data"].get("message"), str)
        self.assertIn("route", payload["data"])

        _, payload = request("GET", "/api/v1/settings/agent")
        self.assert_envelope(payload, True)
        self.assertTrue({"response_style", "personality", "answer_policy"} <= set(payload["data"]))

    def test_06_practice_and_homework_boundaries(self) -> None:
        _, payload = request("POST", f"/api/v1/final-sprint/datasets/{self.dataset_id}/practice", {
            "knowledge_points": ["二重积分"], "question_count": 0, "difficulty": "medium",
        })
        self.assert_envelope(payload, False)
        self.assertEqual(payload["error"]["code"], "INVALID_REQUEST")

        _, payload = request("POST", "/api/v1/homework/hw_missing/hint", {
            "question_id": "q_missing", "user_message": "不知道从哪里开始",
        })
        self.assert_envelope(payload, False)
        self.assertEqual(payload["error"]["code"], "HOMEWORK_NOT_FOUND")


if __name__ == "__main__":
    if not os.getenv("API_BASE_URL"):
        sys.exit("Set API_BASE_URL before running integration tests, e.g. http://localhost:8080")
    unittest.main(verbosity=2)
