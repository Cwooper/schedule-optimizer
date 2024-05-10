// Get references to HTML elements
const courseSelect = document.getElementById('courseSelect'); // Select dropdown for classes
const sectionNumber = document.getElementById('sectionNumber'); // Input field for section number
const courseList = document.getElementById('courseList'); // List element for displaying added classes
let courses = []; // Array to store added classes
let schedule = []; // Array to hold the schedule

function addCourse() {
    const courseName = courseSelect.value; // Get selected class name
    const section = sectionNumber.value; // Get section number
    const errorMessage = document.getElementById('errorMessage'); // Get the error message element

    if (!sectionNumber.checkValidity()) {
        errorMessage.textContent = 'Please enter a valid section number.';
        return;
    }

    if (courseName && section && courses.length < 10) { // Check if all values are provided and total classes are less than 6
        courses.push(`${courseName} ${section}`);
        const li = document.createElement('li'); // Create a new list item element
        li.textContent = `${courseName} ${section}`; // Set text content of the list item
        courseList.appendChild(li); // Append the list item to the classList

        // Create a remove button
        const removeButton = document.createElement('button');
        removeButton.textContent = '-';
        removeButton.addEventListener('click', function() {
            errorMessage.textContent = '';
            courses.splice(courses.indexOf(`${courseName} ${section}`), 1); // Remove class from array
            li.remove(); // Remove list item from the DOM
            // Check if classes array is empty and hide the submit button if it is
            if (courses.length === 0) {
                submitButton.style.display = 'none';
            }
        });

        // Append the remove button to the list item
        li.appendChild(removeButton);

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
        displaySchedule(response);
        // Add to the calendar table
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
    cornerCell.textContent = "Times";
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
    courses.forEach(course => {
        const days = course.days.split(''); // Split the days string into an array
        const startTime = parseInt(course.start_time);
        const endTime = parseInt(course.end_time);
        
        days.forEach(day => {
            const startHour = Math.floor(startTime / 100); // Extract the hour (hundreds place)
            const endHour = Math.ceil(endTime / 100);
            for (let i = startHour; i < endHour; i++) { // Loop through each hour
                const cellId = `${day}-${i}00`; // Append '00' for formatting
                const cell = document.getElementById(cellId);
                const bgColor = stringToColor(course.crn); // Generate color based on CRN
                cell.style.backgroundColor = bgColor; // Set background color
                cell.textContent = `Subject: ${course.subject} Professor: ${course.instructor} GPA: ${course.gpa}`;
                cell.classList.add('scheduled-course'); // Add a class for styling
            }
        });
    });
}

function displaySchedule(response) {
    const firstSchedule = response.schedules[0]; // Assuming you want to display the first schedule
    const courses = firstSchedule.courses;
    addCoursesToCalendar(courses);
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

function handleKeyPress(event) {
    if (event.keyCode === 13) { // Check for pressing Enter
        addCourse();
    }
}

createTable();
