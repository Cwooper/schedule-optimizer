// Get references to HTML elements
const classSelect = document.getElementById('classSelect'); // Select dropdown for classes
const sectionNumber = document.getElementById('sectionNumber'); // Input field for section number
const classList = document.getElementById('classList'); // List element for displaying added classes
let classes = []; // Array to store added classes

// Function to add a class to the list
function addClass() {
    const className = classSelect.value; // Get selected class name
    const section = sectionNumber.value; // Get section number
    if (className && section) { // Check if both class name and section are provided
        classes.push({ className, section }); // Add class to the classes array
        const li = document.createElement('li'); // Create a new list item element
        li.textContent = `${className} - Section ${section}`; // Set text content of the list item
        classList.appendChild(li); // Append the list item to the classList
    }
}

// Function to generate JSON from the classes array
function generateJSON() {
    const json = JSON.stringify(classes); // Convert classes array to JSON string
    console.log(json); // Output JSON string to the console
}


// Fetch the class names from subjects.txt and populate the dropdown menu
fetch('subjects.txt')
    .then(response => response.text())
    .then(text => {
        const classes = text.trim().split('\n');
        classes.forEach(className => {
            const option = document.createElement('option');
            option.value = className;
            option.textContent = className;
            classSelect.appendChild(option);
        });
    })
    .catch(error => {
        console.error('Error fetching subjects:', error);
    });


    document.addEventListener('DOMContentLoaded', function() {
        var calendarEl = document.getElementById('calendar');
      
        var calendar = new FullCalendar.Calendar(calendarEl, {
          timeZone: 'UTC',
          initialView: 'timeGridFiveDay',
          async:false,
          headerToolbar: {
            left: 'prev,next',
            center: 'title'
          },
          views: {
            timeGridFiveDay: {
              type: 'timeGrid',
              duration: { days: 5 },
              buttonText: '5 day',
              title: 'cal',
            }
          },
          events: 'events.json',
          contentHeight:"auto",
          allDaySlot: false,
        });
        
      
        calendar.render();
      });