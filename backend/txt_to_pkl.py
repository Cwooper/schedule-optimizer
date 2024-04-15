# txt_to_pkl.py
# This file turns text files into pickle files.
#
# Unfortunately, to use this file, you have to copy the text of
# (Ctrl+A, Ctrl+C) WWU Classfinder into a "{term}.txt" file.
#
# You can then use this file by doing:
#   python3 txt_to_pkl.py <your_file.txt>
#
# How it works:
# 1. Uses regex to parse the txt file into objects
# 2. Uses pandas to convert the objects into a dataframe
# 3. Uses pandas to store the dataframe as a pickle file
#

import os
import re
import sys
import pandas as pd

############################### REGEX PATTERNS ###############################

# Section
# Number
# Title
# CRN
# Capacity (Seats)
# Enrolled (Seats)
# Available (Seats)
# Instructor Last Name
# Instructor First Name
# Dates
header_pattern = r'.*?([A-Z\/]{3,4})\s*'\
                r'(\d{3}[A-Z]?)\s*'     \
                r'(.*?)\s*'             \
                r'(\d{5})\s*'           \
                r'([-\d]+)\s*'          \
                r'([-\d]+)\s*'          \
                r'([-\d]+)\s*'          \
                r'(.*?),\s*'            \
                r'(.*?)\s*'             \
                r'(\d{2}\/\d{2}-\d{2}\/\d{2})\s*'

# Delivery Method and Days
# Start Time
# End Time
# am/pm
# Building and Room / Type
# Credits
# Fee (optional)
# Fee Type (optional)
footer_pattern = r'.*?DELIVERY\s*(.*?)' \
                 r'(\d{2}:\d{2})-'      \
                 r'(\d{2}:\d{2})\s*'    \
                 r'(am|pm)\s*'          \
                 r'(.*?)\s+'            \
                 r'([\d]|\d-\d)'        \
                 r'(\s+\$[\d]+\.[\d]+)?\s*'\
                 r'(.*)?'

# Days
# Start Time
# End Time
# am/pm
# Building
# Room
lab_pattern = r'\s*(.*?)\s*'            \
              r'(\d{2}:\d{2})-'         \
              r'(\d{2}:\d{2})\s*'       \
              r'(am|pm)\s*'             \
              r'(.*?)\s*'               \
              r'([\d]+)\s*'      

restrictions_pattern = r'\s*Restrictions:\s*(.*)'

prerequisites_pattern = r'\s*Prerequisites:\s*(.*)'

building_room_pattern = r'\s*([A-Z]+)\s+([A-Z0-9]+)\s*'

################################### OBJECTS ###################################


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
    
    def to_dict(self):
        return {
            'delivery': self.delivery,
            'days': self.days,
            'start': self.start,
            'end': self.end,
            'building': self.building,
            'room': self.room,
            'credits': self.credits,
            'fee': self.fee,
            'fee_type': self.fee_type
        }

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
    
    def to_dict(self):
        return {
            'days': self.days,
            'start': self.start,
            'end': self.end,
            'building': self.building,
            'room': self.room
        }
    
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

    def __repr__(self):
        return f'{self.section} {self.number} | {self.title} | {self.crn} | ' \
               f'{self.cap} | {self.enrl} | {self.avail} | '                  \
               f'{self.instructor_last}, {self.instructor_first} | {self.dates}'

    def to_dict(self):
        course_dict = {
            'section': self.section,
            'number': self.number,
            'title': self.title,
            'crn': self.crn,
            'cap': self.cap,
            'enrl': self.enrl,
            'avail': self.avail,
            'instructor_last': self.instructor_last,
            'instructor_first': self.instructor_first,
            'dates': self.dates,
            'restrictions': self.restrictions,
            'prerequisites': self.prerequisites
        }
        if self.footer:
            course_dict.update(self.footer.to_dict())
        if self.lab:
            course_dict.update(self.lab.to_dict())
        return course_dict

    
################################# EXTRACTION #################################

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
        room = None
        class_credits = match.group(6)
        fee = match.group(7)
        fee_type = match.group(8).strip()

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

        building_room_match = re.search(building_room_pattern, building)
        if building_room_match:
            building = building_room_match.group(1)
            room = building_room_match.group(2)

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

########################### CONVERSION AND PARSING ############################

# Turns the list of courses into a pickle file with pandas
def courses_to_pkl(courses, pkl_file):
    courses_dict_list = [course.to_dict() for course in courses]

    df = pd.DataFrame(courses_dict_list)
    df.to_pickle(pkl_file)
    print(df)

# Converts a text file to a pickle file
# Returns the name of the saved pickle file
def txt_to_pkl(txt_file) -> str:
    with open(txt_file, 'r') as file:
        text = file.read()

    courses = []
    lines = text.split('\n')
    current_course = None

    i = 0
    while i < len(lines):
        line = lines[i]
        # Filter empty lines
        if line.strip() == "" and current_course is not None:
            # This is an empty line, which means that 7 lines of garbage follows
            print(f"Found garbage at {i}")
            i += 8
            continue
        elif re.match(header_pattern, line):
            # Add the current course to the list if it's not the first iteration
            if current_course:
                courses.append(current_course)
            current_course = extract_course(line)
        elif re.match(footer_pattern, line):
            current_course.footer = extract_footer(line)
        elif "DELIVERY" not in line and re.match(lab_pattern, line):
            current_course.lab = extract_lab(line)
        elif "Restrictions:" in line:
            match = re.match(restrictions_pattern, line)
            current_course.restrictions = match.group(1).strip()
        elif "Prerequisites:" in line:
            match = re.match(prerequisites_pattern, line)
            current_course.prerequisites = match.group(1).strip()
        elif "TBA" not in line and "N/A" not in line:                   # Fix Later, Bandaid
            # Line is not empty, but does not match above, must be prereqs
            # Unless it is the first lines, when no course has yet been found
            if current_course:
                if current_course.prerequisites is None:
                    current_course.prerequisites = line.strip()
                else:
                    current_course.prerequisites += " " + line.strip()
        i += 1
    # Append the last course to courses
    courses.append(current_course)
    if courses is None:
        print(f"Error: no courses found in {txt_file}")
        exit(1)

    pkl_file = txt_file.replace(".txt", ".pkl")
    courses_to_pkl(courses, pkl_file)
    return pkl_file

def main():
    arguments = sys.argv
    # Check arguments and usage
    if len(arguments) != 2 or not arguments[1].endswith(".txt"):
        print(f"Usage: python3 {arguments[0]} <input_file.txt>")
        exit(1)

    txt_file = arguments[1]
    pkl_file = txt_to_pkl(txt_file)

    print(f"Succesfully converted {txt_file} to {pkl_file}")

if __name__ == "__main__":
    main()