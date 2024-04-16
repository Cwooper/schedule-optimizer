# fetch_schedules.py

# Custom functions and objects
from Course import Course
from Schedule import Schedule
from generate_schedules import generate_schedules

import re
import os
import sys
import pandas as pd
import matplotlib.pyplot as plt

course_name_pattern = r'[A-Z\/ ]{2,4} \d{3}[A-Z]?'
term = 202420

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')
pickle_file = os.path.join(data_directory, str(term) + ".pkl")

# Return a list of schedules based on the course names
def _get_schedules(course_names: list[str], minimum_size=2, maximum_size=5) -> list['Schedule']:
    cleaned_course_names = []
    # Clean and verify course names
    for course_name in course_names:
        course_name = course_name.strip()
        if re.match(course_name_pattern, course_name):
            cleaned_course_names.append(course_name)
        else:
            print(f"Invalid Course Name: {course_name}")
    
    courses = []
    df = pd.read_pickle(pickle_file)
    for course_name in cleaned_course_names:
        # Filter DataFrame based on course name
        matching_rows = df[df['subject'] == course_name]
        course_dicts = matching_rows.to_dict(orient='records')
        if len(course_dicts) == 0:
            print(f"Course not offered this term: {course_name}")
            continue
        for course_dict in course_dicts:
            courses.append(Course(**course_dict))

    schedules = generate_schedules(courses, min_schedule_size=minimum_size, max_schedule_size=maximum_size)
    return schedules

# Weigh each schedule and return the list of the schedules with scores and weights
def weigh_schedules(schedules: list['Schedule']) -> list['Schedule']:
    ...


# Run main for custom use, otherwise use weigh_schedules
def main():
    if len(sys.argv) < 2:
        print("Usage: python3 weigh_schedules.py \"<course1>\" \"<course2>\" ...")
        return
    
    course_names = sys.argv[1:]
    schedules = _get_schedules(course_names)
    for schedule in schedules:
        print(schedule, end="\n\n")
    schedules = weigh_schedules(course_names)
    
if __name__ == "__main__":
    main()
