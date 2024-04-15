from bs4 import BeautifulSoup
import sys              # Remove
sys.path.append('../')  # Remove
import re
from course import Course

input_file = "CSCI.html"

time_pattern = r'\d{2}:\d{2}-\d{2}:\d{2} (am|pm)'

# Returns a start_time, end_time
# Converts from format: "09:00-01:50 pm"
# to military time: "0900", "1350"
def convert_times(times: str):
    time_parts = times.split(" ")
    am_pm = times[1]
    time_start_end = time_parts[0].split("-")
    start_time = time_start_end[0].replace(":", "")
    end_time = time_start_end[1].replace(":", "")

    # Convert times to military time
    if am_pm == "pm":
        if start_time < end_time:
            start_time += 1200
        end_time += 1200
    
    return start_time, end_time

with open(input_file, "r") as f:
    html_content = f.read()

soup = BeautifulSoup(html_content, "html.parser")

# Find tables with specified criteria
tables = soup.find_all("table")
course_tables = []

# Iterate over the tables, cutting out the title table
for table in tables:
    rows = table.find_all("tr")
    for row in rows:
        cells = row.find_all("td", class_="fieldformatboldtext")
        if cells:
            # At this point, I've found the table that has course info
            course_tables.append(table)
            break

print(f"Found {len(course_tables)} course-like tables.")

courses = []

for course_table in course_tables:
    # Grab the td elements from the course
    td_elements = course_table.find_all('td')
    td_text = td_elements[0].get_text(separator="|", strip=True)
    text_parts = td_text.split('|')

    # Skip header tables if they don't contain subject, title, credits
    if len(text_parts) != 3:
        continue

    # Extract course header info from course_table
    subject = text_parts[0]
    course_name = text_parts[1]                             # Unused
    course_credits = text_parts[2].replace("cr", "")

    # Create prerequisites string if it exists
    prerequisites = ""
    if len(td_elements) > 1:
        for td_element in td_elements[1:]:
            prerequisites += td_element.get_text(separator=' ', strip=True)+" "
        prerequisites = prerequisites.strip()

    # Iterate over table siblings to find courses of this header
    sibling = course_table.find_next_sibling("table")
    last_course = None
    while sibling and sibling not in course_tables:
        # Found a course section
        elements = sibling.find_all('td')

        if len(elements) == 13:
            # Append the last course
            if last_course:
                courses.append(last_course)

            # This is a course
            quarter = elements[0].text.strip()
            crn = elements[1].text.strip()
            days = elements[2].text.strip()
            times = elements[3].text.strip()
            instructor = elements[4].text.strip()
            room = elements[5].text.strip()
            addl_fees = ' '.join(elements[6].text.split())
            cap = elements[7].text.strip()
            enrl = elements[8].text.strip()
            avail = elements[9].text.strip()
            waitlist = elements[10].text.strip()
            restrictions = elements[11].text.strip()
            attributes = elements[12].text.strip()

            # Times format: '09:00-01:50 pm'
            if times and re.search(time_pattern, times):
                start_time, end_time = convert_times(times)
            else:
                start_time = None
                end_time = None

            course = Course(subject=subject,
                            course_credits=course_credits,
                            crn=crn,
                            days=days,
                            start_time=start_time,
                            end_time=end_time,
                            instructor=instructor,
                            room=room,
                            addl_fees=addl_fees,
                            cap=cap,
                            enrl=enrl,
                            avail=avail,
                            waitlist=waitlist,
                            restrictions=restrictions,
                            attributes=attributes,
                            prerequisites=prerequisites)
            
            last_course = course

        elif len(elements) == 12:
            # This is a lab
            lab_days = elements[2].text.strip()
            lab_times = elements[3].text.strip()
            lab_instructor = elements[4].text.strip()       # Unused
            lab_room = elements[5].text.strip()
            # Other elements are unneeded for labs

            if lab_times:
                lab_start_time, lab_end_time = convert_times(lab_times)
            else:
                lab_start_time = None
                lab_end_time = None
            # Add lab elements to the last course
            last_course.lab_days = lab_days
            last_course.lab_start_time = lab_start_time
            last_course.lab_end_time = lab_end_time
            last_course.lab_room = lab_room

        else:
            # Skip non-courses or labs
            sibling = sibling.find_next_sibling()
            continue
        
        sibling = sibling.find_next_sibling("table")
            
    courses.append(last_course)

for course in courses:
    print(course)
    print()