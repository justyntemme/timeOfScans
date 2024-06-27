import json
import logging
import os
from typing import Tuple

from dateutil.parser import isoparse
import pytz
import requests

logging.basicConfig(level=logging.INFO)

TL_URL = os.environ.get("TL_URL")


def extractTimeValues(json_response):
    data = json.loads(json_response)
    time_values = []
    for item in data:
        if "time" in item:
            # Convert the 'time' string to a datetime
            time_value = isoparse(item["time"]).astimezone(pytz.utc)
            # Format the datetime object to (hours and minutes in AM/PM format)
            formatted_time = time_value.strftime("%I:%M %p")
            time_values.append(formatted_time)

    return time_values


def getScans(token: str) -> Tuple[int, str]:
    scanURL = TL_URL + "/api/v1/scans" if TL_URL is not None else exit()
    headers = {
        "accept": "application/json; charset=UTF-8",
        "content-type": "application/json",
        "Authorization": f"Bearer {token}",
    }

    response = requests.get(scanURL, headers=headers, timeout=60, verify=False)
    return (response.status_code, response.text)


def generateCwpToken(accessKey: str, accessSecret: str) -> Tuple[int, str]:
    authURL = f"{TL_URL}/api/v1/authenticate" if TL_URL is not None else exit()

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
    if accessKey is None or accessSecret is None:
        logging.error("Missing PC_IDENTITY or PC_SECRET")
    elif TL_URL is None:
        logging.error("Missing TL_URL")
    responseCode, cwpToken = (
        generateCwpToken(accessKey, accessSecret)
        if accessKey and accessSecret
        else (None, None)
    )

    responseCode, content = getScans(cwpToken) if cwpToken else (exit(9))
    logging.info(responseCode)
    timeValues = extractTimeValues(content)
    print(timeValues)


if __name__ == "__main__":
    main()
