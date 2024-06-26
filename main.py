import json
import csv
import logging
import io
import os
import requests
import numpy as np
from typing import Tuple, Optional
import matplotlib.pyplot as plt
import seaborn as sns
from datetime import datetime

logging.basicConfig(level=logging.INFO)

TL_URL = os.environ.get("TL_URL")


def extract_time_values(json_response):
    """
    Extracts 'time' values from a JSON response and adds them to an array.

    Args:
        json_response (str): JSON response as a string.

    Returns:
        list: List of 'time' values as datetime objects.
    """
    # Parse the JSON response
    data = json.loads(json_response)

    time_values = []

    for item in data:
        if "time" in item:
            # Convert the 'time' string to a datetime object and add it to the list
            time_value = datetime.fromisoformat(item["time"])
            formatted_time = time_value.strftime("%I:%M %p")
            time_values.append(formatted_time)

    return time_values


def getScans(token: str) -> Tuple[int, str]:
    scanURL = TL_URL + "/api/v1/scans"
    headers = {
        "accept": "application/json; charset=UTF-8",
        "content-type": "application/json",
        "Authorization": f"Bearer {token}",
    }

    response = requests.get(scanURL, headers=headers, timeout=60, verify=False)
    return (response.status_code, response.text)


def generateCwpToken(accessKey: str, accessSecret: str) -> Tuple[int, str]:
    authURL = TL_URL + "/api/v1/authenticate"
    headers = {
        "accept": "application/json; charset=UTF-8",
        "content-type": "application/json",
    }
    body = {"username": accessKey, "password": accessSecret}
    response = requests.post(
        authURL, headers=headers, json=body, timeout=60, verify=False
    )

    if response.status_code == 200:
        data = json.loads(response.text)
        logging.info("Token acquired")
        return 200, data["token"]
    else:
        logging.error(
            "Unable to acquire token with error code: %s", response.status_code
        )

    return response.status_code, ""


def main():
    accessKey = os.environ.get("PC_IDENTITY")
    accessSecret = os.environ.get("PC_SECRET")
    responseCode, cwpToken = generateCwpToken(accessKey, accessSecret)
    logging.info(responseCode)
    responseCode, content = getScans(cwpToken)
    logging.info(responseCode)
    timeValues = extract_time_values(content)
    # jsonContent = parseString(content)
    print(timeValues)


if __name__ == "__main__":
    main()
