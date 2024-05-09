// Get references to HTML elements
const classSelect = document.getElementById('classSelect'); // Select dropdown for classes
const sectionNumber = document.getElementById('sectionNumber'); // Input field for section number
const classList = document.getElementById('classList'); // List element for displaying added classes
let classes = []; // Array to store added classes
let schedule = []; //array to hold the schedule

function addClass() {
  const className = classSelect.value; // Get selected class name
  const section = sectionNumber.value; // Get section number
  const errorMessage = document.getElementById('errorMessage'); // Get the error message element

  if (className && section && classes.length < 10) { // Check if all values are provided and total classes are less than 6
    classes.push(`${className} ${section}`);
    const li = document.createElement('li'); // Create a new list item element
    li.textContent = `${className} ${section}`; // Set text content of the list item
    classList.appendChild(li); // Append the list item to the classList

    // Create a remove button
    const removeButton = document.createElement('button');
    removeButton.textContent = '-';
    removeButton.addEventListener('click', function() {
      errorMessage.textContent = '';
      classes.splice(classes.indexOf(`${className} ${section}`), 1); // Remove class from array
      li.remove(); // Remove list item from the DOM
      // Check if classes array is empty and hide the submit button if it is
      if (classes.length === 0) {
        submitButton.style.display = 'none';
      }
    });

    // Append the remove button to the list item
    li.appendChild(removeButton);

    // Show the submit button if classes array is not empty
    if (classes.length > 0) {
      submitButton.style.display = 'block';
    }

    // Clear any previous error message
    errorMessage.textContent = '';
  } else {
    // Display an error message if all values are not provided or total classes exceed 6
    errorMessage.textContent = 'Please select a class and section number, and ensure you have less than 10 classes.';
  }
}




// Function to generate JSON from the classes array
function generateJSON() {
  const min = minSelect.value; // Get min value
  const max = maxSelect.value; // Get max value
  const term = termSelect.value; // Get term value
  const quarter = quarterSelect.value; // Get quarter valueif (className && section && min && max && term && quarter) { // Check if all values are provided
    const scheduleinfo = {
        courses: classes,
        min: min,
        max: max,
        term: term + quarter
    };
  schedule= (scheduleinfo);
  const json = JSON.stringify(schedule); // Convert classes array to JSON string
  console.log(json); // Output JSON string to the console
  //give to cwooper
  document.getElementById('jsonDisplay').textContent = JSON.stringify(json, null, 2);

  
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


    