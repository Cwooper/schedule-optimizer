// Get references to HTML elements
const courseSelect = document.getElementById('courseSelect'); // Select dropdown for classes
const sectionNumber = document.getElementById('sectionNumber'); // Input field for section number
const courseList = document.getElementById('courseList'); // List element for displaying added classes

let courses = []; // Array to store courses
let forceList = []; // Array to store forced courses
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
    forceButton.textContent = 'Force';
    forceButton.addEventListener('click', function () {
      if (forceList.includes(`${courseName} ${section}`)) {
        forceList.splice(courses.indexOf(`${courseName} ${section}`), 1);
        forceButton.classList.remove('forced');
      } else {
        forceList.push(`${courseName} ${section}`);
        forceButton.classList.add('forced');
      }
    });
    // Create a remove button
    const removeButton = document.createElement('button');
    removeButton.textContent = '–';
    removeButton.addEventListener('click', function () {
      errorMessage.textContent = '';
      courses.splice(courses.indexOf(`${courseName} ${section}`), 1); // Remove class from array
      forceList.splice(courses.indexOf(`${courseName} ${section}`), 1);
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

  if (forceList.length > max) {
    errorMessage.innerHTML += "Cannot have more forced courses than maximum courses in a schedule.";
    return;
  }

  const scheduleinfo = {
    Courses: courses,
    Forced: forceList,
    Min: parseInt(min),
    Max: parseInt(max),
    Term: term + quarter
  };

  schedule = (scheduleinfo);
  const json = JSON.stringify(schedule); // Convert classes array to JSON string
  console.log(json); // Output JSON string to the console

  fetch('/schedule-optimizer/', {
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
      if (response["Errors"] != null) {
        const errorMessage = document.getElementById('errorMessage');
        errorMessage.innerHTML = response["Errors"];
        if (response["Warnings"] != null) {
          const errorMessage = document.getElementById('errorMessage');
          errorMessage.innerHTML += "<br>" + response["Warnings"].join("<br>");
        }
      } else {
        if (response["Warnings"] != null) {
          const errorMessage = document.getElementById('errorMessage');
          errorMessage.innerHTML += "<br>" + response["Warnings"].join("<br>");
        }
        all_schedules = response["Schedules"]
        sortSchedules();    // Update the sort based on the current value
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

function addCoursesToCalendar(schedule) {
  const courses = schedule.Courses;
  const cornerCell = document.getElementById('cornerCell');
  cornerCell.textContent = `Schedule ${current_schedule}`
  cornerCell._mouseover = () => addHoverEffect.call(cornerCell, "lightgray");
  cornerCell.addEventListener('mouseover', cornerCell._mouseover);
  cornerCell._mouseout = () => removeHoverEffect.call(cornerCell, "lightgray");
  cornerCell.addEventListener('mouseout', cornerCell._mouseout);
  cornerCell._click = () => displaySchedulePopupHandler(schedule);
  cornerCell.addEventListener('click', cornerCell._click);

  courses.forEach(course => {
    const days = course.days.split(''); // Split the days string into an array
    const startTime = parseInt(course.start_time);
    const endTime = parseInt(course.end_time);

    days.forEach(day => {
      let startHour = Math.floor(startTime / 100); // Extract the hour (hundreds place)
      let endHour = Math.ceil(endTime / 100);
      for (let i = startHour; i < endHour; i++) { // Loop through each hour
        if (i < 10) {
          i = `0${i}`
        }
        const cellId = `${day}-${i}00`; // Append '00' for formatting
        const cell = document.getElementById(cellId);
        const bgColor = stringToColor(course.CRN); // Generate color based on CRN
        // Add hover event listener
        cell._mouseover = () => addHoverEffect.call(cell, bgColor);
        cell.addEventListener('mouseover', cell._mouseover);

        cell._mouseout = () => removeHoverEffect.call(cell, bgColor);
        cell.addEventListener('mouseout', cell._mouseout);

        cell._click = () => displayPopupHandler(course);
        cell.addEventListener('click', cell._click);

        cell.style.backgroundColor = bgColor; // Set background color
        cell.innerHTML = `<b>${course.Subject}</b><br>${course.instructor}<br>${course.CRN}<br>${course.room}`;
        cell.classList.add('scheduled-course'); // Add a class for styling
      }
    });

    if (course.lab_days) {
      const labDays = course.lab_days.split('');
      const labStart = parseInt(course.lab_start_time);
      const labEnd = parseInt(course.lab_end_time);
      labDays.forEach(labDay => {
        let labStartHour = Math.floor(labStart / 100); // Extract the hour (hundreds place)
        let labEndHour = Math.ceil(labEnd / 100);
        for (let i = labStartHour; i < labEndHour; i++) { // Loop through each hour
          if (i < 10) {
            i = `0${i}`
          }
          const cellId = `${labDay}-${i}00`; // Append '00' for formatting
          const cell = document.getElementById(cellId);
          const bgColor = stringToColor(course.CRN); // Generate color based on CRN
          // Add hover event listener
          cell._mouseover = () => addHoverEffect.call(cell, bgColor);
          cell.addEventListener('mouseover', cell._mouseover);

          cell._mouseout = () => removeHoverEffect.call(cell, bgColor);
          cell.addEventListener('mouseout', cell._mouseout);

          cell._click = () => displayPopupHandler(course);
          cell.addEventListener('click', cell._click);

          cell.style.backgroundColor = bgColor; // Set background color
          cell.innerHTML = `<b>${course.Subject} LAB</b><br>${course.instructor}<br>${course.CRN}<br>${course.lab_room}`;
          cell.classList.add('scheduled-course'); // Add a class for styling
        }
      });
    }
  });
}

function displaySchedule(schedule) {
  clearSchedule();
  addCoursesToCalendar(schedule);
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

function course_to_str(course) {
  let result = `Instructor: ${course.instructor}<br>
    CRN: ${course.CRN}<br>`;

  result += course.GPA ? `Average GPA: ${course.GPA}<br>` : '';

  result += `Credits: ${course.Credts}<br>
    Room: ${course.room}<br>`

  result += course.days ? `Days: ${course.days}<br>
    Times: ${course.start_time} - ${course.end_time}<br>` : '';

  result += course.lab_days ? `Lab Days: ${course.lab_days}<br>
    Lab Times: ${course.lab_start_time} - ${course.lab_end_time}<br>` : '';

  result += `Available Seats: ${course.AvailableSeats}<br>
    Max Students: ${course.Capacity}<br>
    Students Enrolled: ${course.Enrolled}<br>`

  result += course.WaitlistCount ? `Waitlist: ${course.WaitlistCount}<br>` : '';
  result += course.Prerequisites ? `Prerequisites: ${course.Prerequisites}<br>` : '';
  result += course.Attributes ? `Attributes: ${course.Attributes}<br>` : '';
  result += course.AdditionalFees ? `Additional Fees: ${course.AdditionalFees}` : '';
  result += course.Restrictions ? `Restrictions: ${course.Restrictions}` : '';

  if (result.endsWith('<br>')) {
    // Remove '<br>' from the end of result
    result = result.slice(0, -4);
  }

  return result;
}

function displayHelpPopup() {
  const helpPopup = document.getElementById('help-popup');
  helpPopup.classList.add('active');
}

function closeHelpPopup() {
  const helpPopup = document.getElementById('help-popup');
  helpPopup.classList.remove('active');
}


function displayPopup(course) {
  const oldPopup = document.getElementById('coursePopup');
  if (oldPopup) {
    oldPopup.remove();
  }
  // Create a div element for the popup
  const popup = document.createElement('div');
  popup.id = 'coursePopup';
  popup.classList.add('popup');

  // Add popup header basics
  const popupHeader = document.createElement('div');
  popupHeader.classList.add('popup-header');
  const popupHeaderTitle = document.createElement('div');
  popupHeaderTitle.classList.add('popup-header-title');
  const popupCloseButton = document.createElement('button');
  popupCloseButton.classList.add('popup-close-button');
  popupCloseButton.innerHTML = '&times;'
  popupCloseButton.onclick = () => {
    if (popup) {
      popup.remove();
    }
    return;
  }

  const popupBody = document.createElement('div');
  popupBody.classList.add('popup-body');

  popupHeaderTitle.innerHTML = `<b>${course.Subject}</b>`
  // Add content to the popup
  popupBody.innerHTML = `${course_to_str(course)}`;

  // Append the popup parts together and add it to the document
  popupHeader.appendChild(popupHeaderTitle);
  popupHeader.appendChild(popupCloseButton);
  popup.appendChild(popupHeader);
  popup.appendChild(popupBody)
  document.body.appendChild(popup);
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

function displaySchedulePopupHandler(schedule) {
  displaySchedulePopup(schedule);
}

function displaySchedulePopup(schedule) {
  const oldPopup = document.getElementById('schedulePopup');
  if (oldPopup) {
    oldPopup.remove();
  }
  // Create a div element for the schedulePopup
  const schedulePopup = document.createElement('div');
  schedulePopup.id = 'schedulePopup';
  schedulePopup.classList.add('popup');

  // Add popup header basics
  const popupHeader = document.createElement('div');
  popupHeader.classList.add('popup-header');
  const popupHeaderTitle = document.createElement('div');
  popupHeaderTitle.classList.add('popup-header-title');
  const popupCloseButton = document.createElement('button');
  popupCloseButton.classList.add('popup-close-button');
  popupCloseButton.innerHTML = '&times;'
  popupCloseButton.onclick = () => {
    if (schedulePopup) {
      schedulePopup.remove();
    }
    return;
  }

  const popupBody = document.createElement('div');
  popupBody.classList.add('popup-body');

  popupHeaderTitle.innerHTML = `<b>Schedule ${current_schedule} Weights</b>`
  // Add content to the popup
  popupBody.innerHTML = `Average Score: ${schedule.Score}<br>
                           Average GPA: ${schedule.Weights.GPA * 4.0}<br>
                           Start Time Score: ${schedule.Weights.start}<br>
                           End Time Score: ${schedule.Weights.end}<br>
                           Gaps Score: ${schedule.Weights.gap}`;

  // Append the popup parts together and add it to the document
  popupHeader.appendChild(popupHeaderTitle);
  popupHeader.appendChild(popupCloseButton);
  schedulePopup.appendChild(popupHeader);
  schedulePopup.appendChild(popupBody)
  document.body.appendChild(schedulePopup);
}

function sortSchedules() {
  if (all_schedules.length <= 0) {
    return;
  }

  const sortButton = document.getElementById('scheduleSort');
  const sortValue = sortButton.value;

  all_schedules.sort((a, b) => {
    if (sortValue === 'score') {
      return b.Score - a.Score;
    } else if (sortValue === 'end') {
      return b.Weights.end - a.Weights.end;
    } else if (sortValue === 'gap') {
      return b.Weights.gap - a.Weights.gap;
    } else if (sortValue === 'gpa') {
      return b.Weights.GPA - a.Weights.GPA;
    } else if (sortValue === 'start') {
      return b.Weights.start - a.Weights.start;
    }
    return 0;
  });
}

function updateSort() {
  if (all_schedules.length <= 0) {
    return;
  }
  sortSchedules();
  current_schedule = 0;
  displaySchedule(all_schedules[current_schedule]);
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
