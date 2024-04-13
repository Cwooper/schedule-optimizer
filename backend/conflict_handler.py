from course import Course
from itertools import combinations

def all_conflicts(courses: list[Course]):
    conflicts = []
    for i in range(len(courses)):
        for j in range(i + 1, len(courses)):
            if courses[i].conflicts(courses[j]):
                conflicts.append((courses[i], courses[j]))

    return conflicts

def all_schedules(courses, conflicts, max_schedule_size):
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
            if valid_schedule:
                schedules.append(course_combination)
    return schedules

courses = [
    Course("Math 1", "MTWRF", 900, 1030),
    Course("Math 2", "TR", 1100, 1230),
    Course("Math 3", "MWF", 1200, 1400),
    Course("Math 4", "MTWRF", 800, 820),
    Course("English 1", "MW", 1000, 1130),
    Course("English 2", "TWF", 900, 1030),
    Course("Biology 1", "MWF", 930, 1100),
    Course("Chemistry 1", "TR", 1400, 1530),
    Course("Physics 1", "MWF", 1030, 1200),
    Course("History 1", "MW", 1100, 1230),
    Course("Math 1", "MWF", 800, 900),  # Duplicate name with Math 1
    Course("English 3", "TR", 1200, 1330),
    Course("Chemistry 2", "MTWRF", 830, 1000),
    Course("Biology 2", "TR", 930, 1100),
    Course("Physics 2", "MW", 1030, 1200)
]


conflicts = all_conflicts(courses)
all = all_schedules(courses, conflicts, 5)

for schedule in all:
    print(schedule)
    print("------------")

