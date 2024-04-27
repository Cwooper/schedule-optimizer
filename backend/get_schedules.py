# fetch_schedules.py

# Custom functions and objects
from Course import Course
from Schedule import Schedule
from schedule_generator import generate_schedules

import re
import os
import argparse
import pandas as pd

SUGGESTED_MINIMUM = 2
SUGGESTED_MAXIMUM = 4

course_name_pattern = r'[A-Z\/ ]{2,4} \d{3}[A-Z]?'

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')

usage = """
Usage: python3 get_schedules.py <"course1" "course2" ...> <--term term> [--minimum num] [--maximum num]

Description:
    Retrieve schedules for the specified courses.

Arguments:
    <"course1" "course2" ...>   List of course codes (enclosed in double quotes) to retrieve schedules for.
    <--term term>               Specify the term for which schedules should be retrieved.
    [--minimum num]             (Optional) Specify the minimum number of schedules to retrieve.
    [--maximum num]             (Optional) Specify the maximum number of schedules to retrieve.
"""


# Return a list of schedules based on the course names
def _get_schedules(course_names: list[str], term, minimum_size=2, maximum_size=5) -> list['Schedule']:
    cleaned_course_names = []
    # Clean and verify course names
    for course_name in course_names:
        course_name = course_name.strip()
        if re.match(course_name_pattern, course_name):
            cleaned_course_names.append(course_name)
        else:
            print(f"Invalid Course Name: {course_name}")
    
    courses = []

    pickle_file = os.path.join(data_directory, term, term + '.pkl')
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

def parse_args():
    parser = argparse.ArgumentParser(description="Weigh schedules with optional minimum and maximum values")
    parser.add_argument("courses", nargs="+", help="List of course names")
    parser.add_argument("--term", type=str, default=None, help="Term to schedule")
    parser.add_argument("--minimum", type=int, default=None, help="Minimum size of schedules")
    parser.add_argument("--maximum", type=int, default=None, help="Maximum size of schedules")
    return parser.parse_args()

# Timer function for finding how long it takes for execution
# import time
# def timer(func, *args, **kwargs):
#     start_time = time.time()
#     result = func(*args, **kwargs)
#     end_time = time.time()
#     execution_time = end_time - start_time
#     return result, execution_time

# Run main for custom use, otherwise use weigh_schedules
def main():
    args = parse_args()
    term = args.term
    course_names = args.courses
    minimum_size = args.minimum if args.minimum else SUGGESTED_MINIMUM
    maximum_size = args.maximum if args.maximum else SUGGESTED_MAXIMUM
    
    if not term:
        print(usage)
        return
    if not course_names:
        print("Please provide at least one course name.")
        return
    if len(course_names) < minimum_size:
        print("Please input more courses than your minimum.")
        return 
    
    schedules = _get_schedules(course_names, term, minimum_size=minimum_size, maximum_size=maximum_size)

    # schedules = weigh_schedules(course_names) # TODO

    # schedules, exec_time = timer(_get_schedules, course_names, term, minimum_size=minimum_size, maximum_size=maximum_size)
    # print(f"exec time: {exec_time} seconds")
    
    for schedule in schedules:
        schedule.weigh_self()
        print(schedule, end="\n\n")

if __name__ == "__main__":
    main()
