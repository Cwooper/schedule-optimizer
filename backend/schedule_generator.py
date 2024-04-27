# generate_schedules.py

from Course import Course
from Schedule import Schedule
from itertools import combinations

HARD_MINIMUM = 1
HARD_MAXIMUM = 5
SUGGESTED_MINIMUM = 2
SUGGESTED_MAXIMUM = 4

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

# Return all possible schedules from a list of courses
def generate_schedules(courses: list['Course'], 
                       min_schedule_size=SUGGESTED_MINIMUM,
                       max_schedule_size=SUGGESTED_MAXIMUM) -> list[Schedule]:
    """ Generate schedules based on a list of courses. """
    # Handle possible bounds errors
    if min_schedule_size < HARD_MINIMUM:
        min_schedule_size = HARD_MINIMUM
    if max_schedule_size > HARD_MAXIMUM:
        max_schedule_size = HARD_MAXIMUM
    if min_schedule_size > max_schedule_size:
        min_schedule_size = max_schedule_size
    if len(courses) < min_schedule_size:
        print(f"Cannot generate schedules with {len(courses)} courses.")

    # Find all conflicts between the courses, and return all possible schedules
    conflicts = all_conflicts(courses)
    schedules = all_schedules(courses, conflicts, min_schedule_size, max_schedule_size)
    return schedules
