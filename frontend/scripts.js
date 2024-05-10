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
        document.getElementById('jsonDisplay').textContent = JSON.stringify(response);
    })
    .catch(error => {
        // Handle any errors
        console.error('Error:', error);
    });

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
