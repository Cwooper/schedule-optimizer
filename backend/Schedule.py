# Schedule.py
# Schedule Object
from Course import Course
ROUND = 2

class Schedule:
    def __init__(self, courses: list['Course'], score=None, weights=None):
        self.courses = courses
        self.score = score
        self.weights = weights

    def __repr__(self):
        result = f"score={self.score} | weights={self.weights}\n"
        result += "Subject   Credits CRN    Instructor                GPA   Days  Begin  End   Lab   Begin  End   Building\n"
        for course in self.courses:
            subject = course.subject
            course_credits = course.course_credits
            crn = course.crn
            instructor = course.instructor
            days = course.days
            start = course.start_time
            end = course.end_time
            lab_days = course.lab_days
            lab_start = course.lab_start_time
            lab_end = course.lab_end_time
            gpa = course.gpa
            room = course.room
            
            # Handling None values for lab related attributes
            lab_days_str = lab_start_str = lab_end_str = ""
            if lab_days is not None:
                lab_days_str = f"{lab_days:<5}"
            if lab_start is not None:
                lab_start_str = f"{lab_start:<4}"
            if lab_end is not None:
                lab_end_str = f"{lab_end:<4}"

            course_string = f"{subject:<9} {course_credits:7} {crn:6} {instructor:<25} {round(gpa, ROUND):<5} {days:5} {start:6} "\
                            f"{end:5} {lab_days_str:5} {lab_start_str:6} {lab_end_str:5} {room:6}\n"
            result += course_string
        
        return result
    
    def _weigh_gpa(self):
        # GPA 
        total_gpa = sum(course.gpa for course in self.courses)
        course_num = len(self.courses)
        
        # Calculate the average GPA, normalize it to a score between 0 and 1
        average_gpa = total_gpa / course_num
        gpa_score = average_gpa / 4.0   # Weigh based on 4.0 scale     
        return round(gpa_score, ROUND)
    
    def _weigh_start(self):
        start_time = min(int(course.start_time) for course in self.courses)
        start_time = to_mins(start_time)

        if start_time < 480 or start_time >= 780:  # Before 08:00 or after/equal to 13:00
            return 0.0
        elif 480 <= start_time < 600: # 0800 to 1000
            start_score = (start_time - 480) / 120.0
            return round(start_score, ROUND)
        elif 600 <= start_time < 660: # 1000 to 1100
            return 1.0
        else:
            start_score = (780 - start_time) / 120.0 # 1100 to 1300
            return round(start_score, ROUND) 
        
    def _weigh_end(self):
        end_time = max(int(course.end_time) for course in self.courses)
        end_time = to_mins(end_time)
        
        if end_time <= 840: # 1400 in mins from midnight
            return 1.0
        elif end_time > 960: # 1600 in mins from midnight
            return 0.0
        else:
            end_score = (960 - end_time) / 120.0 # Between 1400 and 1600
            return round(end_score, ROUND)
        
    def _weigh_gaps(self):
        end_time = max(int(course.end_time) for course in self.courses)
        end_time_mins = to_mins(end_time)
        start_time = min(int(course.start_time) for course in self.courses)
        start_time_mins = to_mins(start_time)

        day_time_mins = end_time_mins - start_time_mins
        
        course_time_mins = 0
        for course in self.courses:
            course_start = to_mins(course.start_time)
            course_end = to_mins(course.end_time)
            course_time_mins += (course_end - course_start)
        
        gap_time = day_time_mins - course_time_mins

        num_courses = len(self.courses)
        best_gaps_time = 10 * (num_courses - 1)

        if gap_time <= best_gaps_time:
            return 1.0
        elif gap_time > (120 + best_gaps_time):
            return 0.0
        else:
            # y=mx+b from best_gaps_time (1.0) to 120+best_gaps_time (0.0)
            gap_score = ((-1 / 120) * (gap_time - best_gaps_time)) + 1.0
            return round(gap_score, ROUND)

    def weigh_self(self):
        gpa_score = self._weigh_gpa()
        start_score = self._weigh_start()
        end_score = self._weigh_end()
        gap_score = self._weigh_gaps()

        START_WEIGHT = 2
        GPA_WEIGHT = 5
        END_WEIGHT = 2
        GAP_WEIGHT = 1

        WEIGHT_TOTAL = START_WEIGHT + END_WEIGHT + GAP_WEIGHT + GPA_WEIGHT

        score =  (end_score * END_WEIGHT +
                  start_score * START_WEIGHT +
                  gap_score * GAP_WEIGHT +
                  gpa_score * GPA_WEIGHT) / WEIGHT_TOTAL
        
        self.weights = {
            "start": start_score,
            "end": end_score,
            "gap": gap_score,
            "gpa": gpa_score
        }
        self.score = round(score, ROUND)

# Returns minutes since midnight
def to_mins(time: str) -> int:
    time = int(time)
    return (time // 100) * 60 + (time % 100)
