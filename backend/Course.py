# Course.py
# Course object

class Course:
    def __init__(self, subject=None,  course_credits=None, crn=None, days=None,
                 start_time=None,     end_time=None,      lab_days=None,
                 lab_start_time=None, lab_end_time=None,  instructor=None,
                 room=None,           lab_room=None,      addl_fees=None, 
                 cap=None, enrl=None, avail=None,         waitlist=None,
                 restrictions=None,   attributes=None,    prerequisites=None):
        self.subject = subject
        self.course_credits = course_credits
        self.crn = crn
        self.days = days
        self.start_time = start_time
        self.end_time = end_time
        self.lab_days = lab_days
        self.lab_start_time = lab_start_time
        self.lab_end_time = lab_end_time
        self.instructor = instructor
        self.room = room
        self.lab_room = lab_room
        self.addl_fees = addl_fees
        self.cap = cap
        self.enrl = enrl
        self.avail = avail
        self.waitlist = waitlist
        self.restrictions = restrictions
        self.attributes = attributes
        self.prerequisites = prerequisites

    def __repr__(self):
        return (
            f"Course(subject={self.subject}, course_credits={self.course_credits}, "
            f"crn={self.crn}, days={self.days}, start_time={self.start_time}, "
            f"end_time={self.end_time}, lab_days={self.lab_days}, "
            f"lab_start_time={self.lab_start_time}, lab_end_time={self.lab_end_time}, "
            f"instructor={self.instructor}, room={self.room}, addl_fees={self.addl_fees}, "
            f"cap={self.cap}, enrl={self.enrl}, avail={self.avail}, "
            f"waitlist={self.waitlist}, restrictions={self.restrictions}, "
            f"attributes={self.attributes}, prerequisites={self.prerequisites})"
        )

    def to_dict(self):
        return {
            "subject": self.subject,
            "course_credits": self.course_credits,
            "crn": self.crn,
            "days": self.days,
            "start_time": self.start_time,
            "end_time": self.end_time,
            "lab_days": self.lab_days,
            "lab_start_time": self.lab_start_time,
            "lab_end_time": self.lab_end_time,
            "instructor": self.instructor,
            "room": self.room,
            "lab_room": self.lab_room,
            "addl_fees": self.addl_fees,
            "cap": self.cap,
            "enrl": self.enrl,
            "avail": self.avail,
            "waitlist": self.waitlist,
            "restrictions": self.restrictions,
            "attributes": self.attributes,
            "prerequisites": self.prerequisites
        }
    
    def conflicts(self, other: 'Course') -> bool:
        # Two courses confilict if they have the same subject
        if self.subject == other.subject:
            return True
        # Don't include TBD or N/A courses in schedules to save runtime
        if self.days in ("TBD", "N/A") or other.days in ("TBD", "N/A"):
            return True
        # Compare self days with the others days
        for day in self.days:
            if day in other.days:
                if (_time_conflict(self.start_time, self.end_time,
                                   other.start_time, other.end_time)):
                    return True
        # Compare self's lab with other's days
        if self.lab_days:
            for day in self.lab_days:
                if day in other.days:
                    if (_time_conflict(self.lab_start_time, self.lab_end_time,
                                       other.start_time, other.end_time)):
                        return True
        # Compare other's lab with self's days
        if other.lab_days:
            for day in other.lab_days:
                if day in self.days:
                    if (_time_conflict(self.start_time, self.end_time,
                                    other.lab_start_time, other.lab_end_time)):
                        return True
        # Compare lab times between both labs
        if self.lab_days and other.lab_days:
            for day in self.lab_days:
                if day in other.lab_days:
                    if (_time_conflict(self.lab_start_time, self.lab_end_time,
                                    other.lab_start_time, other.lab_end_time)):
                        return True
        return False

# Time conflict helper method
def _time_conflict(start1, end1, start2, end2):
    return (((start1 >= start2) and (start1 <= end2)) or
            ((end1 >= start2) and (end1 <= end2)))
