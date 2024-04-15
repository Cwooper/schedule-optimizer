import os
import requests
from datetime import datetime, timedelta
from bs4 import BeautifulSoup
from course import Course

MAX_SUBJECT_WAIT =  30      # Days
MAX_COURSE_WAIT =   2       # Days

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')
subjects_file = os.path.join(data_directory, 'subjects.txt')

url = 'https://web4u.banner.wwu.edu/pls/wwis/wwskcfnd.TimeTable'

def _fetch_subjects_from_url() -> list:
    # Fetch from server if MAX_SUBJECT_WAIT since last update
    print(f"Longer than {MAX_SUBJECT_WAIT} days since last update, "
          f"fetching from server.")
    r = requests.get(url)

    if r.status_code != 200:
        print(f"Error: status code {r.status_code} accessing {url}")
        exit(1)

    # Find the select subject element and store them into a subjects list
    soup = BeautifulSoup(r.text, 'html.parser')
    select_element = soup.find('select', id='subj')
    subjects = [option['value'] for option in select_element.find_all('option')]

    # Format subjects into file for later use
    with open(subjects_file, 'w') as file:
        file.write(datetime.now().isoformat() + '\n')
        for subject in subjects:
            file.write(subject + '\n')

    print("Successfully fetched from server.")
    return subjects

# Fetches subjects from the main url
def fetch_subjects_list() -> list:
    # Check if subjects file exists and is newer than one day
    if os.path.exists(subjects_file):
        with open(subjects_file, 'r') as file:
            lines = file.read().split('\n')
            file_time = datetime.fromisoformat(lines[0])
            max_wait = timedelta(days=MAX_SUBJECT_WAIT)

            if datetime.now() - file_time < max_wait:
                print(f"Shorter than {MAX_SUBJECT_WAIT} days since last "
                      f"update, retrieving subjects file.")
                # Cut off the timestamp and newline at the end
                return lines[1:-1]
    else:
        return _fetch_subjects_from_url()

# Fetches courses from url if they're not stored in a file
def _fetch_courses_from_url(subject: str, subject_file: str) -> list['Course']:
    payload = {
        'term': 202420, 
        'curr_yr': 2324,
        'subj': subject
    }

    r = requests.post(url, data=payload)
    if r.status_code != 200:
        print(f"Error: status code {r.status_code} accessing {url}")
        exit(1)

    courses = []
    soup = BeautifulSoup(r.text, 'html.parser')
    with open(subject_file, 'w', encoding='utf-8') as f:
        f.write(soup.prettify())
        print(f"Wrote {subject_file} from response.text")
    # Find all td elements with class "fieldformatboldtext"
    # course_elements = soup.find_all("td", class_="fieldformatboldtext")
    # 
    # for course in course_elements:
    #     # Extracting name, title, credits, and prerequisites
    #     lines = course.get_text().split("\n")
    #     name = lines[0].strip()
    #     title = lines[1].strip()
    #     credits = lines[2].strip()
    #     prerequisites = "\n".join(lines[3:]).strip()  # Combine remaining lines for prerequisites
    #     print("Name:", name)
    #     print("Title:", title)
    #     print("Credits:", credits)
    #     print("Prerequisites:", prerequisites)
    #     print()

def _file_to_courses(lines: list[str]) -> list['Course']:
    courses = []
    for line in lines:
        # Parse the line
        courses.extend(...)
    return courses

# Returns a list of courses provided by one subject
def fetch_courses(subject: str) -> list['Course']:
    # Check if subjects file exists and is newer than one day
    subject_file = os.path.join(data_directory, subject + '.txt')
    if os.path.exists(subject_file):
        with open(subject_file, 'r') as file:
            lines = file.read().split('\n')
            file_time = datetime.fromisoformat(lines[0])
            max_wait = timedelta(days=MAX_COURSE_WAIT)

            if datetime.now() - file_time < max_wait:
                print(f"Shorter than {MAX_COURSE_WAIT} days since last "
                      f"update, retrieving subjects file.")
                # Cut off the timestamp and newline at the end

                return _file_to_courses(lines[1:-1])
    else:
        return _fetch_courses_from_url(subject, subject_file)


def main():
    subjects = fetch_subjects_list()
    print(subjects)
    # courses = []
    #for subject in subjects:
    fetch_courses(subjects[23])
    # courses.extend(fetch_courses(subjects[0]))

if __name__ == "__main__":
    main()
