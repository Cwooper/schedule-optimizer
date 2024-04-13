import pandas as pd
import re


header_pattern = r'([A-Z\/]{4})\s*'     \
                r'(\d{3}[A-Z]?)\s*'     \
                r'(.*?)\s*'             \
                r'(\d{5})\s*'           \
                r'([-\d]+)\s*'          \
                r'([-\d]+)\s*'          \
                r'([-\d]+)\s*'          \
                r'(.*?),\s*'            \
                r'(.*?)\s*'             \
                r'(\d{2}\/\d{2}-\d{2}\/\d{2})\s*'

footer_pattern = r'\s*DELIVERY\s*(.*?)\s*'\
                 r'(\d{2}:\d{2})-'      \
                 r'(\d{2}:\d{2})\s*'    \
                 r'(am|pm)\s*'          \
                 r'([A-Z]+)\s*'         \
                 r'([A-Z\d]+)\s*'       \
                 r'([\d]|\d-\d)\s*'     \
                 r'(\$\d{2}\.\d{2})?\s*'\
                 r'(.*?)'

lab_pattern = r'\s*(.*?)\s*'            \
              r'(\d{2}:\d{2})-'         \
              r'(\d{2}:\d{2})\s*'       \
              r'(am|pm)\s*'             \
              r'(.*?)\s*'               \
              r'([\d]+)\s*'      

restrictions_pattern = r'\s*Restrictions:\s*(.*)'

prerequisites_pattern = r'\s*Prerequisites:\s*(.*)'


class Footer:
    def __init__(self, delivery, days, start, end, building, room, credits,
                 fee, fee_type):
        self.delivery = delivery
        self.days = days
        self.start = start
        self.end = end
        self.building = building
        self.room = room
        self.credits = credits
        self.fee = fee
        self.fee_type = fee_type
    
    def __repr__(self):
        return f'{self.delivery} | {self.days} | {self.start} - {self.end} | '\
               f'{self.building} {self.room} | {self.credits} | ' \
               f'{self.fee} {self.fee_type}'

class Lab:
    def __init__(self, days, start, end, building, room):
        self.days = days
        self.start = start
        self.end = end
        self.building = building
        self.room = room

    def __repr__(self):
        return f'{self.days} | {self.start} - {self.end} | {self.building} '  \
               f'{self.room}'
    
class Course:
    def __init__(self, section, number, title, crn, cap, enrl, avail, 
                 instructor_last, instructor_first, dates, footer=None,
                 lab=None, restrictions=None, prerequisites=None):
        self.section = section
        self.number = number
        self.title = title
        self.crn = crn
        self.cap = cap
        self.enrl = enrl
        self.avail = avail
        self.instructor_last = instructor_last
        self.instructor_first = instructor_first
        self.dates = dates

        self.footer = footer
        self.lab = lab
        self.restrictions = restrictions
        self.prerequisites = prerequisites

    def add_footer(self, footer):
        self.footer = footer
    
    def add_lab(self, lab):
        self.lab = lab
    
    def add_restrictions(self, restrictions):
        self.restrictions = restrictions

    def add_prerequisites(self, prequisites):
        self.prerequisites = prequisites
    
    def __repr__(self):
        return f'{self.section} {self.number} | {self.title} | {self.crn} | ' \
               f'{self.cap} | {self.enrl} | {self.avail} | '                  \
               f'{self.instructor_last}, {self.instructor_first} | {self.dates}'

def extract_course(line: str) -> Course:
    match = re.search(header_pattern, line)
    if match:
        section = match.group(1)
        number = match.group(2)
        title = match.group(3)
        crn = match.group(4)
        cap = match.group(5)
        enrl = match.group(6)
        avail = match.group(7)
        instructor_last = match.group(8)
        instructor_first = match.group(9)
        dates = match.group(10)
        return Course(section, number, title, crn, cap, enrl, avail, instructor_last, instructor_first, dates)
    else:
        return None

def extract_footer(line: str) -> Footer:
    match = re.search(footer_pattern, line)
    if match:
        delivery_days = match.group(1).strip()
        delivery = None
        days = None
        start_time = match.group(2)
        end_time = match.group(3)
        am_pm = match.group(4)
        building = match.group(5).strip()
        room = match.group(6)
        class_credits = match.group(7)
        fee = match.group(8)
        fee_type = match.group(9)

        start_time = int(start_time.replace(":", ""))
        end_time = int(end_time.replace(":", ""))

        if am_pm == "pm":
            if start_time < end_time:
                start_time += 1200
            end_time += 1200

        split_string = delivery_days.split()

        if any(day in split_string[-1] for day in ['M', 'T', 'W', 'R', 'F']):
            days = split_string[-1]
            delivery = ' '.join(split_string[:-1])

        return Footer(delivery, days, start_time, end_time, building, room,
                      class_credits, fee, fee_type)
    else:    
        return None

def extract_lab(line: str) -> Lab:
    match = re.search(lab_pattern, line)
    if match:
        days = match.group(1)
        start_time = match.group(2)
        end_time = match.group(3)
        am_pm = match.group(4)
        building = match.group(5)
        room = match.group(6)

        start_time = int(start_time.replace(":", ""))
        end_time = int(end_time.replace(":", ""))

        if am_pm == "pm":
            if start_time < end_time:
                start_time += 1200
            end_time += 1200
        
        return Lab(days, start_time, end_time, building, room)
    else:
        return None
    


# Load the HTML data
with open("courses.txt", "r", encoding="utf-8") as file:
    page_text = file.read()

lines = page_text.split('\n')

courses = []

# Check if TBA or N/A on line before adding fooer, regex won't catch it

for line in lines:
    if line:
        footer = extract_footer(line)
        if footer:
            courses.append(footer)
    
for course in courses:
    print(course)
