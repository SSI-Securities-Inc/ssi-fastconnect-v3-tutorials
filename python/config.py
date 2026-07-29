"""
config.py — Tải cấu hình tự động từ config.json ở thư mục gốc (nếu có)
========================================================================
Module này tự động đọc tệp `config.json` tại gốc dự án (ssi-fastconnect-v3-tutorials/config.json).
Nếu không có `config.json`, nó sẽ dùng giá trị mặc định/placeholder.
"""

import json
import os
from ssi_sdk import Config

CONFIG_PATH = os.path.join(os.path.dirname(__file__), "..", "config.json")


def load_config():
    if os.path.exists(CONFIG_PATH):
        try:
            with open(CONFIG_PATH, "r", encoding="utf-8") as f:
                data = json.load(f)
            cfg = Config(
                client_id=data.get("client_id", "<CLIENT_ID>"),
                api_key=data.get("api_key", "<API_KEY>"),
                api_secret=data.get("api_secret", "<API_SECRET>"),
                private_key=data.get("private_key", "<PRIVATE_KEY_CONTENT>"),
                log_level=data.get("log_level", "INFO"),
            )
            account_no = data.get("equity_account", "<ACCOUNT_NO>")
            otp = data.get("otp", "<OTP>")
            return cfg, account_no, otp
        except Exception as e:
            print(f"Warning: Không thể đọc file config.json: {e}")

    return (
        Config(
            client_id="<CLIENT_ID>",
            api_key="<API_KEY>",
            api_secret="<API_SECRET>",
            private_key="<PRIVATE_KEY_CONTENT>",
            log_level="DEBUG",
        ),
        "<ACCOUNT_NO>",
        "<OTP>",
    )


config, ACCOUNT_NO, OTP = load_config()
