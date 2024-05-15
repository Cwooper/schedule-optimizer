# generate_schedules.py

import os
import re
import pandas as pd

from backend.models.Course import Course
from backend.models.Schedule import Schedule
from itertools import combinations

HARD_MINIMUM = 1
HARD_MAXIMUM = 6

course_name_pattern = r'[A-Z\/ ]{2,4} \d{3}[A-Z]?'

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')

# Takes a list if course names, verifies, cleans them, and returns the result
def clean_course_names(course_names):
    response = {
        "cleaned_course_names": [],
        "warnings": [],
        "errors": []
    }

    cleaned_course_names = []
    # Clean and verify course names
    for course_name in course_names:
        course_name = course_name.strip()
        if re.match(course_name_pattern, course_name):
            cleaned_course_names.append(course_name)
        else:
            response["warnings"].append(f"Invalid Course Name: {course_name}")

    response["cleaned_course_names"] = cleaned_course_names
    return response

def all_conflicts(courses: list[Course]):
    conflicts = []
    for i in range(len(courses)):
        for j in range(i + 1, len(courses)):
            if courses[i].conflicts(courses[j]):
                conflicts.append((courses[i], courses[j]))

    return conflicts

def all_schedules(courses, conflicts, min_schedule_size, max_schedule_size) -> list['Schedule']:
    course_combinations = []
    # For every permutation size
    for size in range(min_schedule_size, max_schedule_size + 1):
        # Iterate through every combination in permutations
        for course_combination in combinations(courses, size):
            valid_schedule = True

            # Check if conflict pair is in the current combination
            for conflict_pair in conflicts:
                if all(course in course_combination for course in conflict_pair):
                    valid_schedule = False
                    break

            # Add it to all schedules if it is valid
            if valid_schedule:
                course_combinations.append(Schedule(list(course_combination)))

    return course_combinations

# Finds courses based on a list of course_names and a term
def get_courses(course_names: list[str], term: str):
    result = {
        "warnings": [],
        "errors": [],
        "courses": []
    }
    response = clean_course_names(course_names)
    cleaned_course_names = response["cleaned_course_names"]
    if response["warnings"] != []:
        result["warnings"].append(response["warnings"])
    if response["errors"] != []:
        result["errors"].append(response["errors"])

    courses = []

    pickle_file = os.path.join(data_directory, term, term + '.pkl')
    try:
        df = pd.read_pickle(pickle_file)
    except Exception as e:
        print(e)
        result["errors"].append(f"Could not find {term}")
        return result

    for course_name in cleaned_course_names:
        # Filter DataFrame based on course name
        matching_rows = df[df['subject'] == course_name]
        course_dicts = matching_rows.to_dict(orient='records')
        if len(course_dicts) == 0:
            result["warnings"].append(f"Course not offered this term: {course_name}")
            continue
        for course_dict in course_dicts:
            courses.append(Course(**course_dict))

    result["courses"] = courses
    return result

# Return all possible schedules from a list of courses
def generate_schedules(courses: list['Course'], 
                       min_schedule_size,
                       max_schedule_size):
    """ Generate schedules based on a list of courses. """

    response = {
        "schedules": [],
        "warnings": [],
        "errors": []
    }
    # Handle possible bounds errors
    if min_schedule_size < HARD_MINIMUM:
        min_schedule_size = HARD_MINIMUM
    if max_schedule_size > HARD_MAXIMUM:
        max_schedule_size = HARD_MAXIMUM
    if min_schedule_size > max_schedule_size:
        min_schedule_size = max_schedule_size
    if len(courses) < min_schedule_size:
        response["errors"].append(f"Cannot generate schedules with {len(courses)} courses.")
        return response

    # Find all conflicts between the courses, and return all possible schedules
    conflicts = all_conflicts(courses)
    schedules = all_schedules(courses, conflicts, min_schedule_size, max_schedule_size)
    response["schedules"] = schedules
    return response
