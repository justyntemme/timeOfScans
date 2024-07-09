import json
import os
from typing import Tuple, List
from dateutil.parser import isoparse
import pytz
import requests
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed

# Global Variables
n = None  # To shorten line lengths
tlUrl = os.environ.get("TL_URL")

if tlUrl is None:
    raise ValueError("TL_URL environment variable is not set")


def getAllScansWithTimeCounts(token: str, limit: int = 100) -> dict:
    """
    Fetches all scans from the API, extracts time values, and counts the occurrences of each time.

    Args:
        token (str): The authorization token.
        limit (int): Number of items per page.

    Returns:
        dict: A dictionary with times as keys and the number of occurrences as values.
    """
    offset = 0
    timeCounts = Counter()
    while True:
        status_code, response_text = getScans(token, offset, limit)
        if status_code != 200:
            print(f"Error fetching scans: {status_code}")
            break

        # Debugging: Print the response text
        print(
            f"Response text: {response_text[:200]}..."
        )  # Print the first 200 characters for brevity

        # Extract time values using the provided function
        try:
            timeValues = extractTimeValues(response_text)
            print(
                f"Extracted time values: {timeValues}"
            )  # Debugging: Print extracted time values
            timeCounts.update(timeValues)
        except ValueError as e:
            print(f"Error extracting time values: {e}")
            break

        # Check if we have reached the end of the data
        try:
            data = json.loads(response_text)
            if not isinstance(data, list) or len(data) < limit:
                break
        except json.JSONDecodeError as e:
            print(f"Error decoding JSON response: {e}")
            break

        # Update the offset to get the next page of data
        offset += limit

    return dict(timeCounts)


def getScans(token: str, offset: int = 0, limit: int = 100) -> Tuple[int, str]:
    """
    Fetches scans from the API with pagination.

    Args:
        token (str): The authorization token.
        offset (int): The offset from where to start fetching the results.
        limit (int): The number of results to return.

    Returns:
        Tuple[int, str]: A tuple containing the
        response status code and response text.
    """
    scanUrl = tlUrl + "/api/v1/scans" if tlUrl is not None else exit(1)
    headers = {
        "accept": "application/json; charset=UTF-8",
        "content-type": "application/json",
        "Authorization": f"Bearer {token}",
    }

    # Define the query parameters for pagination
    params = {"offset": offset, "limit": limit}

    response = requests.get(
        scanUrl, headers=headers, params=params, timeout=60, verify=False
    )
    return (response.status_code, response.text)


def generateCwpToken(accessKey: str, accessSecret: str) -> Tuple[int, str]:
    """
    Generates a CWP token using the provided access key and secret.

    Args:
        accessKey (str): The access key.
        accessSecret (str): The access secret.

    Returns:
        Tuple[int, str]: A tuple containing
        the response status code and the token.
    """
    authUrl = f"{tlUrl}/api/v1/authenticate" if tlUrl is not n else exit(1)

    headers = {
        "accept": "application/json; charset=UTF-8",
        "content-type": "application/json",
    }
    body = {"username": accessKey, "password": accessSecret}
    response = requests.post(
        authUrl, headers=headers, json=body, timeout=60, verify=False
    )

    if response.status_code == 200:
        data = json.loads(response.text)
        print("Token acquired")
        return 200, data["token"]
    else:
        raise ValueError(
            f"Unable to acquire token with error code: {response.status_code}"
        )


def checkParam(paramName: str) -> str:
    """
    Checks if a parameter is set in the environment variables.

    Args:
        paramName (str): The name of the parameter.

    Returns:
        str: The value of the parameter.

    Raises:
        ValueError: If the parameter is not found.
    """
    paramValue = os.environ.get(paramName)
    if paramValue is None:
        raise ValueError(f"Missing {paramName}")
    return paramValue


def extractTimeValues(jsonResponse: str) -> list:
    """
    Extracts and formats time values from a JSON response.

    Args:
        jsonResponse (str): The JSON response as a string.

    Returns:
        list: A list of formatted time values.
    """
    if not jsonResponse:
        raise ValueError("Empty JSON response")

    # Check if the response is HTML
    if jsonResponse.strip().startswith("<!doctype html>"):
        raise ValueError("Received HTML response instead of JSON")

    # Check if the response is null
    if jsonResponse.strip() == "null":
        raise ValueError("Received null response")

    try:
        data = json.loads(jsonResponse)
        if not isinstance(data, list):
            print(f"Unexpected JSON structure: {data}")
            raise ValueError("Expected a list in the JSON response")
    except json.JSONDecodeError as e:
        raise ValueError(f"Error decoding JSON response: {e}")

    timeValues = []
    for item in data:
        if "entityInfo" in item and "scanTime" in item["entityInfo"]:
            try:
                # Convert the 'scanTime' string to a datetime
                timeValue = isoparse(item["entityInfo"]["scanTime"]).astimezone(
                    pytz.utc
                )
                # Format the datetime object to (hours and minutes in AM/PM format)
                formattedTime = timeValue.strftime("%I:%M %p")
                timeValues.append(formattedTime)
            except Exception as e:
                print(f"Error parsing time value: {e}")

    return timeValues


def main():
    """
    Main function to execute the script.
    """
    params: Tuple[str, str, str] = ("PC_IDENTITY", "PC_SECRET", "TL_URL")
    accessKey, accessSecret, _ = map(checkParam, params)
    responseCode, cwpToken = (
        generateCwpToken(accessKey, accessSecret)
        if accessKey and accessSecret
        else (None, None)
    )
    print(responseCode)
    timeCounts = getAllScansWithTimeCounts(cwpToken, 5) if cwpToken else (exit(1))
    print(timeCounts)


if __name__ == "__main__":
    main()
