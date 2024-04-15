# Course object

class Course:
    def __init__(self, subject=None,  class_credits=None, crn=None, days=None,
                 start_time=None,     end_time=None,      lab_days=None,
                 lab_start_time=None, lab_end_time=None,  instructor=None,
                 room=None,           lab_room=None,      addl_fees=None, 
                 cap=None, enrl=None, avail=None,         waitlist=None,
                 restrictions=None,   attributes=None,    prerequisites=None):
        self.subject = subject
        self.class_credits = class_credits
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
            f"Course(subject={self.subject}, class_credits={self.class_credits}, "
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
            "class_credits": self.class_credits,
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
        # Compare self days with the others days
        for day in self.days:
            if day in other.days:
                if (((self.start_time >= other.start_time) and 
                     (self.start_time <= other.end_time)) or
                     ((self.end_time >= other.start_time) and 
                      (self.end_time <= other.end_time))):
                    return True
        # Compare self's lab with other's days
        if self.lab_days:
            for day in self.lab_days:
                if day in other.days:
                    if (((self.lab_start_time >= other.start_time) and 
                         (self.lab_start_time <= other.end_time)) or
                        ((self.lab_end_time >= other.start_time) and 
                         (self.lab_end_time <= other.end_time))):
                        return True
        # Compare other's lab with self's days
        if other.lab_days:
            for day in other.lab_days:
                if day in self.days:
                    if (((self.start_time >= other.lab_start_time) and 
                         (self.start_time <= other.lab_end_time)) or
                        ((self.end_time >= other.lab_start_time) and 
                         (self.end_time <= other.lab_end_time))):
                        return True
        # Compare lab times between both labs
        if self.lab_days and other.lab_days:
            for day in self.lab_days:
                if day in other.lab_days:
                    if (((self.lab_start_time >= other.lab_start_time) and 
                         (self.lab_start_time <= other.lab_end_time)) or
                        ((self.lab_end_time >= other.lab_start_time) and 
                         (self.lab_end_time <= other.lab_end_time))):
                        return True
        return False
