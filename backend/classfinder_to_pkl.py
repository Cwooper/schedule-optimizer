import os
import requests
from datetime import datetime, timedelta
from bs4 import BeautifulSoup
from course import Course

MAX_SUBJECT_WAIT =  30      # Days
MAX_COURSE_WAIT =   2       # Days

subjects_file = "data/subjects.txt"
terms_file = "data/terms.txt"
pkl_file = "data/"

url = 'https://web4u.banner.wwu.edu/pls/wwis/wwskcfnd.TimeTable'

s = requests.Session()

fetch_sub

def fetch_subjects_list() -> list:
    # Check if subjects file exists and is newer than one day
    if os.path.exists(subjects_file):
        with open(subjects_file, 'r') as file:
            lines = file.read().split('\n')
            file_time = datetime.fromisoformat(lines[0])
            max_wait = timedelta(days=MAX_SUBJECT_WAIT)

            if datetime.now() - file_time < max_wait:
                print(f"Shorter than {MAX_SUBJECT_WAIT} days since last, "    \ 
                      f"update, retrieving subjects file.")
                # Cut off the timestamp and newline
                return lines[1:-1]

    # Fetch from server if MAX_SUBJECT_WAIT since last update
    print(f"Longer than {MAX_SUBJECT_WAIT} days since last update, "          \
          f"fetching from server.")
    response = s.get(url)

    if response.status_code != 200:
        print(f"Error: status code {response.status_code} accessing {url}")
        exit(1)

    soup = BeautifulSoup(response.text, 'html.parser')

    # Find the select subject element
    select_element = soup.find('select', id='subj')

    # Store every subject into a subjects list
    subjects = [option['value'] for option in select_element.find_all('option')]

    # Format subjects into file for later use
    with open(subjects_file, 'w') as file:
        file.write(datetime.now().isoformat() + '\n')
        for subject in subjects:
            file.write(subject + '\n')

    print("Successfully fetched from server.")
    return subjects

# Returns a list of courses provided by one subject
def fetch_subject(subject: str) -> list:


def main():
    subjects = fetch_subjects_list()
    
    for subject in subjects:
        payload = {
            'term': 202420,                           # Edit later
            'curr_yr': 2324,
            'subj': subject
        }
        # TODO Retrieve every file from each subject from the website.
        # Store each in a file system like the above.

if __name__ == "__main__":
    main()
