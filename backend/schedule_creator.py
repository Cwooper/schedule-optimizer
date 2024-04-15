from course import Course
from itertools import combinations

def all_conflicts(courses: list[Course]):
    conflicts = []
    for i in range(len(courses)):
        for j in range(i + 1, len(courses)):
            if courses[i].conflicts(courses[j]):
                conflicts.append((courses[i], courses[j]))

    return conflicts

def all_schedules(courses, conflicts, min_schedule_size, max_schedule_size):
    schedules = []
    # For every permutation size
    for size in range(1, max_schedule_size + 1):
        # Iterate through every combination in permutations
        for course_combination in combinations(courses, size):
            valid_schedule = True

            # Check if conflict pair is in the current combination
            for conflict_pair in conflicts:
                if all(course in course_combination for course in conflict_pair):
                    valid_schedule = False
                    break

            # Add it to all schedules if it is valid
            if valid_schedule and len(course_combination) >= min_schedule_size:
                schedules.append(course_combination)

    return schedules

courses = [
    Course(subject="Math 1", days="MTWRF", start_time=900, end_time=1030),
    Course(subject="Math 2", days="TR", start_time=1100, end_time=1230),
    Course(subject="Math 3", days="MWF", start_time=1200, end_time=1400),
    Course(subject="Math 4", days="MTWRF", start_time=800, end_time=820),
    Course(subject="Math 1", days="TR", start_time=1300, end_time=1350, lab_days="TW", lab_start_time=1150, lab_end_time=1250)
    # Course(subject="English 1", days="MW", start_time=1000, end_time=1130),
    # Course(subject="English 2", days="TWF", start_time=900, end_time=1030),
    # Course(subject="Biology 1", days="MWF", start_time=930, end_time=1100),
    # Course(subject="Chemistry 1", days="TR", start_time=1400, end_time=1530),
    # Course(subject="Physics 1", days="MWF", start_time=1030, end_time=1200),
    # Course(subject="History 1", days="MW", start_time=1100, end_time=1230),
    # Course(subject="Math 1", days="MWF", start_time=800, end_time=900)
]


conflicts = all_conflicts(courses)
all = all_schedules(courses, conflicts, 2, 3)

for schedule in all:
    print(schedule)
    print("------------")

