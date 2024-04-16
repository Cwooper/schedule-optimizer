# Schedule.py
# Schedule Object
from Course import Course

class Schedule:
    def __init__(self, courses: list['Course'], score=None, weights=None):
        self.courses = courses
        self.score = score
        self.weights = weights

    def __repr__(self):
        result = f"score={self.score} | weights={self.weights}\n"
        for course in self.courses:
            subject = course.subject
            crn = course.crn
            instructor = course.instructor
            days = course.days
            start = course.start_time
            end = course.end_time
            lab_days = course.lab_days
            lab_start = course.lab_start_time
            lab_end = course.lab_end_time
            
            # Handling None values for lab related attributes
            lab_days_str = lab_start_str = lab_end_str = ""
            if lab_days is not None:
                lab_days_str = f"{lab_days:<5}"
            if lab_start is not None:
                lab_start_str = f"{lab_start:<4}"
            if lab_end is not None:
                lab_end_str = f"{lab_end:<4}"

            course_string = f"{subject:<19} {crn:5} {instructor:<20} {days:<5} {start:<4} "\
                            f"{end:<4} {lab_days_str:<5} {lab_start_str:<4} {lab_end_str:<4}\n"
            result += course_string
        
        return result