import os
import re
import json
import requests
import pandas as pd
from bs4 import BeautifulSoup
from datetime import datetime, timedelta

# Custom course object
from course import Course

MAX_SUBJECT_WAIT =  30      # Days
MAX_COURSE_WAIT =   2       # Days

time_pattern = r'\d{2}:\d{2}-\d{2}:\d{2} (am|pm)'

term = 202420
year = 2324

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')
subjects_file = os.path.join(data_directory, 'subjects.txt')
pickle_file = os.path.join(data_directory, str(term) + ".pkl")


url = 'https://web4u.banner.wwu.edu/pls/wwis/wwskcfnd.TimeTable'

# Converts times format from 
# "09:00-01:50 pm" to "0900", "1350"
def convert_times(times: str):
    time_parts = times.split(" ")                   # "09:00-13:50", "pm"
    am_pm = time_parts[1]                           # "pm"
    time_start_end = time_parts[0].split("-")       # "09:00", "01:50"
    start_time = time_start_end[0].replace(":", "") # "0900"
    end_time = time_start_end[1].replace(":", "")   # "0150"

    # Convert times to military time
    if am_pm == "pm":
        if start_time < end_time:
            # Convert to and from int for time conversions
            start_time = int(start_time)
            start_time += 1200
            start_time = str(start_time)

        # Convert to and from int for time conversions
        end_time = int(end_time)                    # "0150"
        end_time += 1200
        end_time = str(end_time)                    # "1350"
    
    return start_time, end_time                     # "0900", "1350"

# Finds all of the course-like tables and returns them
def _get_course_tables(tables):
    # Iterate over the tables, cutting out the title table
    course_tables = []
    for table in tables:
        rows = table.find_all("tr")
        for row in rows:
            cells = row.find_all("td", class_="fieldformatboldtext")
            if cells:
                # At this point, I've found the table that has course info
                course_tables.append(table)
                break
    return course_tables

# Extracts and returns a course from the given element
def _extract_course(elements, subject, course_credits, prerequisites) -> Course:
    # This is a course
    quarter = elements[0].text.strip()              # Unused
    crn = elements[1].text.strip()
    days = elements[2].text.strip()
    times = elements[3].text.strip()
    instructor = elements[4].text.strip()
    room = elements[5].text.strip()
    addl_fees = ' '.join(elements[6].text.split())
    cap = elements[7].text.strip()
    enrl = elements[8].text.strip()
    avail = elements[9].text.strip()
    waitlist = elements[10].text.strip()
    restrictions = elements[11].text.strip()
    attributes = elements[12].text.strip()

    # Times format: '09:00-01:50 pm'
    if times and re.search(time_pattern, times):
        start_time, end_time = convert_times(times)
    else:
        start_time = times
        end_time = times

    course = Course(subject=subject,
                    course_credits=course_credits,
                    crn=crn,
                    days=days,
                    start_time=start_time,
                    end_time=end_time,
                    instructor=instructor,
                    room=room,
                    addl_fees=addl_fees,
                    cap=cap,
                    enrl=enrl,
                    avail=avail,
                    waitlist=waitlist,
                    restrictions=restrictions,
                    attributes=attributes,
                    prerequisites=prerequisites)
    return course


# Returns all the given courses from the given course tables
def _get_courses(course_tables):
    courses = []
    for course_table in course_tables:
        # Grab the td elements from the course
        td_elements = course_table.find_all('td')
        td_text = td_elements[0].get_text(separator="|", strip=True)
        text_parts = td_text.split('|')

        # Skip header tables if they don't contain subject, title, credits
        if len(text_parts) != 3:
            continue

        # Extract course header info from course_table
        subject = text_parts[0]
        course_name = text_parts[1]                             # Unused
        course_credits = text_parts[2].replace("cr", "")

        # Create prerequisites string if it exists
        prerequisites = ""
        if len(td_elements) > 1:
            for td_element in td_elements[1:]:
                prerequisites += td_element.get_text(separator=' ', strip=True)+" "
            prerequisites = prerequisites.strip()

        # Iterate over table siblings to find courses of this header
        sibling = course_table.find_next_sibling("table")
        last_course = None
        while sibling and sibling not in course_tables:
            # Found a course section
            elements = sibling.find_all('td')

            if len(elements) == 13:
                # Append the last course
                if last_course:
                    courses.append(last_course)

                last_course = _extract_course(elements, subject,
                                              course_credits, prerequisites)

            elif len(elements) == 12:
                # This is a lab
                lab_days = elements[2].text.strip()
                lab_times = elements[3].text.strip()
                lab_instructor = elements[4].text.strip()       # Unused
                lab_room = elements[5].text.strip()
                # Other elements are unneeded for labs

                if lab_times and re.search(time_pattern, lab_times):
                    lab_start_time, lab_end_time = convert_times(lab_times)
                else:
                    lab_start_time = lab_times
                    lab_end_time = lab_times
                # Add lab elements to the last course
                last_course.lab_days = lab_days
                last_course.lab_start_time = lab_start_time
                last_course.lab_end_time = lab_end_time
                last_course.lab_room = lab_room

            else:
                # Skip non-courses or labs
                sibling = sibling.find_next_sibling()
                continue
            
            sibling = sibling.find_next_sibling("table")
            
        # Append the very last course
        courses.append(last_course)
    return courses

# Fetches all subjects list from classfinder url
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

# Finds the course file text and returns a list of Courses
def _file_to_courses(subject_file: str) -> list['Course']:
    courses = []
    with open(subject_file, 'r') as file:
        page_text = file.read()

        # Remove the first line 
        newline_idx = page_text.find('\n')
        page_text = page_text[newline_idx + 1:]

        course_dicts = json.loads(page_text)
        for course_dict in course_dicts:
            courses.append(Course(**course_dict))
        return courses

# Creates a courses file
def _courses_to_file(subject_file: str, courses: list['Course']):
    current_datetime = datetime.now().isoformat()
    courses_json = json.dumps([course.to_dict() for course in courses], indent=4)

    with open(subject_file, 'w') as file:
        file.write(current_datetime + '\n')
        file.write(courses_json)
    
    print(f"Succesfully saved courses into {subject_file}")

# Fetches courses from url if they're not stored in a file
def _fetch_courses_from_url(subject: str, subject_file: str) -> list['Course']:
    payload = {
        'term': term, 
        'curr_yr': year,
        'subj': subject
    }

    r = requests.post(url, data=payload)
    if r.status_code != 200:
        print(f"Error: status code {r.status_code} accessing {url}")
        exit(1)

    soup = BeautifulSoup(r.text, 'html.parser')
    course_tables = _get_course_tables(soup.find_all("table"))

    courses = _get_courses(course_tables)
    # Write the course to a file for later use
    _courses_to_file(subject_file, courses)

    return courses

# Returns a list of courses provided by one subject
def fetch_courses(subject: str) -> list['Course']:
    # Check if subjects file exists and is newer than one day
    cleaned_subject = subject.replace('/', '-')
    subject_file = os.path.join(data_directory, cleaned_subject + '.txt')

    if os.path.exists(subject_file):
        with open(subject_file, 'r') as file:
            lines = file.read().split('\n')
            file_time = datetime.fromisoformat(lines[0])
            max_wait = timedelta(days=MAX_COURSE_WAIT)

            if datetime.now() - file_time < max_wait:
                print(f"Shorter than {MAX_COURSE_WAIT} days since last "
                      f"update, retrieving subjects file.")
                return _file_to_courses(subject_file)
    else:
        return _fetch_courses_from_url(subject, subject_file)
    
def courses_to_pickle(courses: list['Course']):
    course_dicts = [course.to_dict() for course in courses]
    df = pd.DataFrame(course_dicts)
    df.to_pickle(pickle_file)
    print(f"Successfully turned courses into a pickle at {pickle_file}")

def main():
    subjects = fetch_subjects_list()
    print(f"Found {len(subjects)} subjects")
    courses = []
    for subject in subjects:
        new_courses = fetch_courses(subject)
        print(f"{subject}: Found {len(new_courses)} courses")
        courses.extend(new_courses)

    print(f"Found {len(courses)} courses")

    courses_to_pickle(courses)
    print("Program ran succesffully!")

if __name__ == "__main__":
    main()
