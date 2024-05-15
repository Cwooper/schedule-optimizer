// Get references to HTML elements
const courseSelect = document.getElementById('courseSelect'); // Select dropdown for classes
const sectionNumber = document.getElementById('sectionNumber'); // Input field for section number
const courseList = document.getElementById('courseList'); // List element for displaying added classes
let courses = []; // Array to store added classes
let schedule = []; // Array to hold the schedule
let all_schedules = [];
let current_schedule = 0;

function addCourse() {
    const courseName = courseSelect.value; // Get selected class name
    const section = sectionNumber.value; // Get section number
    const errorMessage = document.getElementById('errorMessage'); // Get the error message element
    errorMessage.textContent = '';

    if (courses.includes(`${courseName} ${section}`)) {
        errorMessage.textContent = 'Duplicate Course Entered.';
        return;
    }

    if (!sectionNumber.checkValidity()) {
        errorMessage.textContent = 'Please enter a valid section number.';
        return;
    }

    if (courseName && section && courses.length < 10) { // Check if all values are provided and total classes are less than 6
        courses.push(`${courseName} ${section}`);
        const li = document.createElement('li'); // Create a new list item element
        courseList.appendChild(li); // Append the list item to the classList

        const forceButton = document.createElement('button');
        forceButton.textContent = 'F';
        forceButton.addEventListener('click', function() {
            console.log("forced")
        });
        // Create a remove button
        const removeButton = document.createElement('button');
        removeButton.textContent = '–';
        removeButton.addEventListener('click', function() {
            errorMessage.textContent = '';
            courses.splice(courses.indexOf(`${courseName} ${section}`), 1); // Remove class from array
            li.remove(); // Remove list item from the DOM
            // Check if classes array is empty and hide the submit button if it is
            if (courses.length === 0) {
                submitButton.style.display = 'none';
            }
        });

        const textContainer = document.createElement('div');
        textContainer.textContent = `${courseName} ${section}`;
        textContainer.style.flex = '1';
        li.appendChild(textContainer)
        li.appendChild(forceButton);
        li.appendChild(removeButton);
        li.style.display = 'flex';
        li.style.justifyContent = 'space-between';
        // Show the submit button if classes array is not empty
        if (courses.length > 0) {
            submitButton.style.display = 'block';
        }

        // Clear any previous error message
        errorMessage.textContent = '';
    } else {
        // Display an error message if all values are not provided or total classes exceed 6
        errorMessage.textContent = 'Ensure you have less than 10 classes added.';
    }
}

// Function to generate JSON from the classes array
function generateJSON() {
    const min = minSelect.value; // Get min value
    const max = maxSelect.value; // Get max value
    const term = termSelect.value; // Get term value
    const quarter = quarterSelect.value; // Get quarter valueif (className && section && min && max && term && quarter) { // Check if all values are provided
    if (courses.length < min) {
        errorMessage.innerHTML += "Cannot have less courses than minimum courses in a schedule.";
        return;
    }
    
    const scheduleinfo = {
        courses: courses,
        min: min,
        max: max,
        term: term + quarter
    };
    schedule = (scheduleinfo);
    const json = JSON.stringify(schedule); // Convert classes array to JSON string
    console.log(json); // Output JSON string to the console
    // document.getElementById('jsonDisplay').textContent = JSON.stringify(schedule, null, 2);

    fetch('/schedule-optimizer', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: json
    })
    .then(response => response.json())
    .then(response => {
        // Display the response in the console
        console.log(response); // Here is the response
        if (response["errors"].length > 0) {
            const errorMessage = document.getElementById('errorMessage');
            errorMessage.innerHTML = response["errors"];
            if (response["warnings"].length > 0) {
                const errorMessage = document.getElementById('errorMessage');
                errorMessage.innerHTML += "<br>" + response["warnings"].join("<br>");
            }
        } else {
            if (response["warnings"].length > 0) {
                const errorMessage = document.getElementById('errorMessage');
                errorMessage.innerHTML += "<br>" + response["warnings"].join("<br>");
            }
            all_schedules = response["schedules"]
            current_schedule = 0;
            displaySchedule(all_schedules[current_schedule]);
        }
    })
    .catch(error => {
        // Handle any errors
        console.error('Error:', error);
    });

}

function createTable() {
    // Get the table element
    const hours = ['0800', '0900', '1000', '1100', '1200', '1300', '1400', '1500', '1600', '1700'];
    const daysOfWeek = ['M', 'T', 'W', 'R', 'F'];
    const fullDaysOfWeek = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday']
    const table = document.getElementById('calendar');
    
    // Create the header row
    const headerRow = document.createElement('tr');
    
    // Create an empty cell for the corner
    const cornerCell = document.createElement('th');
    cornerCell.id = "cornerCell"
    headerRow.appendChild(cornerCell);
    
    // Create header cells for each day of the week
    fullDaysOfWeek.forEach(day => {
        const headerCell = document.createElement('th');
        headerCell.textContent = day;
        headerRow.appendChild(headerCell);
    });
    
    // Append the header row to the table
    table.appendChild(headerRow);
    
    // Loop through each hour
    hours.forEach(hour => {
      // Create a new table row
      const row = document.createElement('tr');
  
      // Create a table cell for the hour
      const hourCell = document.createElement('td');
      hourCell.textContent = `${hour.slice(0, 2)}:${hour.slice(2)}`;
      row.appendChild(hourCell);
  
      // Loop through each day of the week
      daysOfWeek.forEach(day => {
        // Create a new table cell
        const cell = document.createElement('td');
        cell.id = `${day}-${hour}`;
        row.appendChild(cell);
      });
  
      // Append the row to the table
      table.appendChild(row);
      table.classList.add('calendar-table'); // So this applies after DOM is loaded
    });
}

function stringToColor(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const c = (hash & 0x00FFFFFF)
        .toString(16)
        .toUpperCase()
        .padStart(6, '0');
    return `#${c}`;
}

function addCoursesToCalendar(courses) {
    const cornerCell = document.getElementById('cornerCell');
    cornerCell.textContent = `Schedule ${current_schedule}`
    courses.forEach(course => {
        const days = course.days.split(''); // Split the days string into an array
        const startTime = parseInt(course.start_time);
        const endTime = parseInt(course.end_time);
        
        days.forEach(day => {
            var startHour = Math.floor(startTime / 100); // Extract the hour (hundreds place)
            var endHour = Math.ceil(endTime / 100);
            for (let i = startHour; i < endHour; i++) { // Loop through each hour
                if (i < 10) {
                    i = `0${i}`
                }
                const cellId = `${day}-${i}00`; // Append '00' for formatting
                const cell = document.getElementById(cellId);
                const bgColor = stringToColor(course.crn); // Generate color based on CRN
                // Add hover event listener
                cell._mouseover = () => addHoverEffect.call(cell, bgColor);
                cell.addEventListener('mouseover', cell._mouseover);
                
                cell._mouseout = () => removeHoverEffect.call(cell, bgColor);
                cell.addEventListener('mouseout', cell._mouseout);
                
                cell._click = () => displayPopupHandler(course);
                cell.addEventListener('click', cell._click);

                cell.style.backgroundColor = bgColor; // Set background color
                cell.innerHTML = `<b>${course.subject}</b><br>${course.instructor}<br>${course.crn}<br>${course.room}`;
                cell.classList.add('scheduled-course'); // Add a class for styling
            }
        });

        if(course.lab_days) {
            const labDays = course.lab_days.split('');
            const labStart = parseInt(course.lab_start_time);
            const labEnd = parseInt(course.lab_end_time);
            labDays.forEach(labDay => {
                var labStartHour = Math.floor(labStart / 100); // Extract the hour (hundreds place)
                var labEndHour = Math.ceil(labEnd / 100);
                for (let i = labStartHour; i < labEndHour; i++) { // Loop through each hour
                    if (i < 10) {
                        i = `0${i}`
                    }
                    const cellId = `${labDay}-${i}00`; // Append '00' for formatting
                    const cell = document.getElementById(cellId);
                    const bgColor = stringToColor(course.crn); // Generate color based on CRN
                    // Add hover event listener
                    cell._mouseover = () => addHoverEffect.call(cell, bgColor);
                    cell.addEventListener('mouseover', cell._mouseover);
                    
                    cell._mouseout = () => removeHoverEffect.call(cell, bgColor);
                    cell.addEventListener('mouseout', cell._mouseout);
                    
                    cell._click = () => displayPopupHandler(course);
                    cell.addEventListener('click', cell._click);

                    cell.style.backgroundColor = bgColor; // Set background color
                    cell.innerHTML = `<b>${course.subject} LAB</b><br>${course.instructor}<br>${course.crn}<br>${course.lab_room}`;
                    cell.classList.add('scheduled-course'); // Add a class for styling
                }
            });
        }
    });
}

function displaySchedule(schedule) {
    clearSchedule();
    const courses = schedule.courses;
    addCoursesToCalendar(courses);
}

function clearSchedule() {
    const table = document.getElementById('calendar');
    const elements = table.getElementsByTagName('*');
    
    for (let i = 0; i < elements.length; i++) {
        const element = elements[i];
        if (element.id) {
            element.textContent = ''; // Clear text content
            element.style = ''; // Clear style
            element.classList.remove("scheduled-course");
            element.removeEventListener('mouseover', element._mouseover);
            element.removeEventListener('mouseout', element._mouseout);
            element.removeEventListener('click', element._click);
        }
    }
}

function nextSchedule() {
    if (current_schedule + 1 < all_schedules.length) {
        current_schedule++;
        errorMessage.textContent = "";
        displaySchedule(all_schedules[current_schedule]);
    } else {
        const errorMessage = document.getElementById('errorMessage');
        errorMessage.textContent = "You are at the end of the schedules";
    }
}

function prevSchedule() {
    if (current_schedule - 1 > -1) {
        current_schedule--;
        errorMessage.textContent = "";
        displaySchedule(all_schedules[current_schedule]);
    } else {
        const errorMessage = document.getElementById('errorMessage');
        errorMessage.textContent = "You are at the beginning of the schedules";
    }
}


function displayPopup(course) {
    const oldPopup = document.getElementById('coursePopup');
    if (oldPopup) {
        closePopup();
    }
    // Create a div element for the popup
    const popup = document.createElement('div');
    // Set ID for the popup
    popup.id = 'coursePopup';
    // Set styles for the popup
    popup.classList.add('popup'); // Add a class for styling if need    ed
    popup.style.position = 'fixed';
    popup.style.top = '50%';
    popup.style.left = '50%';
    popup.style.transform = 'translate(-50%, -50%)';
    popup.style.border = '2px solid black'; // or any border style you prefer
    popup.style.padding = '10px'; // Adjust padding as needed
    popup.style.backgroundColor = 'lightblue'; // or any background color you prefer

    // Add content to the popup
    popup.innerHTML = `
        <span style="cursor: pointer; position: absolute; top: 5px; right: 5px; font-weight: bold; font-size: 20px" onclick="closePopup()">X</span>
        <b>${course.subject}</b><br>
        Instructor: ${course.instructor}<br>
        CRN: ${course.crn}<br>
        Average GPA: ${course.gpa}<br>
        Credits: ${course.course_credits}<br>
        Room: ${course.room}<br>
        Days: ${course.days}<br>
        Times: ${course.start_time}-${course.end_time}<br>
        Lab Days: ${course.lab_days}<br>
        Lab Times: ${course.lab_start_time}-${course.lab_end_time}<br>
        Available Seats: ${course.avail}<br>
        Max Students: ${course.cap}<br>
        Students Enrolled: ${course.enrl}<br>
        Waitlist: ${course.waitlist}<br>
        Prerequisites: ${course.prerequisites}<br>
        Attributes: ${course.attributes}<br>
        Additional Fees: ${course.addl_fees}
    `;

    // Append the popup to the document body
    document.body.appendChild(popup);
}

function closePopup() {
    const popup = document.getElementById('coursePopup');
    if (popup) {
        popup.remove();
    }
}

// Clear box and add course
function handleKeyPress(event) {
    if (event.keyCode === 13) { // Check for pressing Enter
        addCourse();
        const sectionNumber = document.getElementById('sectionNumber');
        sectionNumber.value = '';
    }
}

function addHoverEffect(bgColor) {
    this.style.backgroundColor = 'lightblue';
}

function removeHoverEffect(bgColor) {
    this.style.backgroundColor = bgColor;
}

function displayPopupHandler(course) {
    displayPopup(course);
}

// Fetch the class names from subjects.txt and populate the dropdown menu
fetch('subjects.txt')
    .then(response => response.text())
    .then(text => {
        const courses = text.trim().split('\n');
        courses.forEach(courseName => {
            const option = document.createElement('option');
            option.value = courseName;
            option.textContent = courseName;
            courseSelect.appendChild(option);
        });
    })
    .catch(error => {
        console.error('Error fetching subjects:', error);
});

createTable();
