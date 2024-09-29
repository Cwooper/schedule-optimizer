// Get references to HTML elements
const courseSelect = document.getElementById('courseSelect'); // Select dropdown for classes
const sectionNumber = document.getElementById('sectionNumber'); // Input field for section number
const courseList = document.getElementById('courseList'); // List element for displaying added classes

let courses = []; // Array to store courses
let forceList = []; // Array to store forced courses
let schedule = []; // Array to hold the schedule
let asyncCourses = []; // Array to hold async/tbd courses
let all_schedules = [];
let current_schedule = 0;

function addCourse() {
  const courseName = courseSelect.value; // Get selected class name
  const section = sectionNumber.value; // Get section number
  const errorMessage = document.getElementById('errorMessage'); // Get the error message element
  errorMessage.style.display = 'none';

  if (courses.includes(`${courseName} ${section}`)) {
    errorMessage.style.display = 'block';
    errorMessage.textContent = 'Duplicate Course Entered.';
    return;
  }

  if (!sectionNumber.checkValidity()) {
    errorMessage.style.display = 'block';
    errorMessage.textContent = 'Please enter a valid section number.';
    return;
  }

  if (courseName && section && courses.length < 13) { // Check if all values are provided and total classes are less than 13
    courses.push(`${courseName} ${section}`);
    const li = document.createElement('li'); // Create a new list item element
    courseList.appendChild(li); // Append the list item to the classList

    const forceButton = document.createElement('button');
    forceButton.textContent = 'Force';
    forceButton.addEventListener('click', function () {
      if (forceList.includes(`${courseName} ${section}`)) {
        forceList.splice(forceList.indexOf(`${courseName} ${section}`), 1);
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
      errorMessage.style.display = 'none';
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
    errorMessage.style.display = 'none';
  } else {
    // Display an error message if all values are not provided or total classes exceed 6
    errorMessage.style.display = 'block';
    errorMessage.textContent = 'Ensure you have less than 13 classes added.';
  }
}

// Function to generate JSON from the classes array
function generateJSON() {
  let min = minSelect.value; // Get min value
  const max = maxSelect.value; // Get max value
  const term = termSelect.value; // Get term value
  const quarter = quarterSelect.value; // Get quarter valueif (className && section && min && max && term && quarter) { // Check if all values are provided
  if (courses.length < min) { // Lowering minium vaue to 1 if there are less courses than min rather than error
    min = courses.length
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
  console.log(`Sending: ${json}`); // Output JSON string to the console

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
      console.log('Response:', response);

      // Size of JSON (should be less than 2MB ever)
      const jsonString = JSON.stringify(response);
      const sizeInBytes = new Blob([jsonString]).size;
      console.log("JSON size:", sizeInBytes, "bytes");

      // Clear async courses if their previously were some.
      if (asyncCourses && asyncCourses.length > 0) {
        clearAsyncCourses();
      }

      const errorMessage = document.getElementById('errorMessage');
      let messages = [];

      if (response.Errors && response.Errors.length > 0) {
        messages = messages.concat(response.Errors);
      }

      if (response.Warnings && response.Warnings.length > 0) {
        messages = messages.concat(response.Warnings);
      }

      if (messages.length > 0) {
        errorMessage.innerHTML = messages.join("<br>");
        errorMessage.style.display = 'block';
      } else {
        errorMessage.style.display = 'none';
      }

      asyncCourses = response.Asyncs;
      if (asyncCourses && asyncCourses.length > 0) {
        displayAsyncCourses();
      }

      if (response.Schedules) {
        all_schedules = response.Schedules;
        sortSchedules();    // Update the sort based on the current value
        current_schedule = 0;
        displaySchedule(all_schedules[current_schedule]);
      }
    })
    .catch(error => {
      // Handle any errors
      console.error('Error:', error);
      const errorMessage = document.getElementById('errorMessage');
      errorMessage.innerHTML = 'An error occurred while processing your request.';
      errorMessage.style.display = 'block';
    });
}

function createTable() {
  // Get the table element
  const hours = ['0800', '0900', '1000', '1100', '1200', '1300', '1400', '1500', '1600', '1700', '1800'];
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
  cornerCell.textContent = `Schedule ${current_schedule + 1} / ${all_schedules.length}`
  cornerCell._mouseover = () => addHoverEffect.call(cornerCell, "lightgray");
  cornerCell.addEventListener('mouseover', cornerCell._mouseover);
  cornerCell._mouseout = () => removeHoverEffect.call(cornerCell, "lightgray");
  cornerCell.addEventListener('mouseout', cornerCell._mouseout);
  cornerCell._click = () => displaySchedulePopupHandler(schedule);
  cornerCell.addEventListener('click', cornerCell._click);

  courses.forEach(course => {
    course.Sessions.forEach(session => {
      if (session.IsAsync || session.IsTimeTBD) {
        return; // Skip async or TBD sessions
      }

      const days = session.Days.split('');
      const startTime = session.StartTime;
      const endTime = session.EndTime;

      days.forEach(day => {
        let startHour = Math.floor(startTime / 100);
        let endHour = Math.ceil(endTime / 100);
        for (let i = startHour; i < endHour; i++) {
          if (i < 10) {
            i = `0${i}`
          }
          const cellId = `${day}-${i}00`;
          const cell = document.getElementById(cellId);
          const bgColor = stringToColor(course.CRN.toString());

          cell._mouseover = () => addHoverEffect.call(cell, bgColor);
          cell.addEventListener('mouseover', cell._mouseover);

          cell._mouseout = () => removeHoverEffect.call(cell, bgColor);
          cell.addEventListener('mouseout', cell._mouseout);

          cell._click = () => displayPopupHandler(course, session);
          cell.addEventListener('click', cell._click);

          cell.style.backgroundColor = bgColor;
          cell.innerHTML = `<b>${course.Subject}</b><br>${session.Instructor}<br>${course.CRN}<br>${session.Location}`;
          cell.classList.add('scheduled-course');
        }
      });
    });
  });
}

function displaySchedule(schedule) {
  clearSchedule();
  addCoursesToCalendar(schedule);
  if (asyncCourses && asyncCourses.length > 0) {
    displayAsyncCourses();
  }
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
    errorMessage.style.display = 'none';
    displaySchedule(all_schedules[current_schedule]);
  } else {
    const errorMessage = document.getElementById('errorMessage');
    errorMessage.style.display = 'block';
    errorMessage.textContent = "You are at the end of the schedules";
  }
}

function prevSchedule() {
  if (current_schedule - 1 > -1) {
    current_schedule--;
    errorMessage.style.display = 'none';
    displaySchedule(all_schedules[current_schedule]);
  } else {
    const errorMessage = document.getElementById('errorMessage');
    errorMessage.style.display = 'block';
    errorMessage.textContent = "You are at the beginning of the schedules";
  }
}

function course_to_str(course, session) {
  let result = `Course: ${course.Subject} - ${course.Title}<br>
    CRN: ${course.CRN}<br>`;

  result += course.GPA ? `Average GPA: ${Number(course.GPA).toFixed(2)}<br>` : '';

  result += `Credits: ${course.Credits}<br><br>`;

  result += session.IsAsync ? 'Asynchronous Session<br>' :
    (session.IsTimeTBD ? 'Session Time TBD<br>' :
      `Days: ${session.Days}<br>
             Times: ${session.StartTime.toString().padStart(4, '0')} - ${session.EndTime.toString().padStart(4, '0')}<br>`);
  result += `Instructor: ${session.Instructor}<br>
             Location: ${session.Location}<br><br>`;

  result += `Available Seats: ${course.AvailableSeats}<br>
    Max Students: ${course.Capacity}<br>
    Students Enrolled: ${course.Enrolled}<br>`;

  result += course.Prerequisites ? `Prerequisites: ${course.Prerequisites}<br>` : '';
  result += course.Attributes ? `Attributes: ${course.Attributes}<br>` : '';
  result += course.AdditionalFees ? `Additional Fees: ${course.AdditionalFees}<br>` : '';
  result += course.Restrictions ? `Restrictions: ${course.Restrictions}` : '';

  if (result.endsWith('<br>')) {
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


function displayPopup(course, session) {
  const oldPopup = document.getElementById('coursePopup');
  if (oldPopup) {
    oldPopup.remove();
  }
  const popup = document.createElement('div');
  popup.id = 'coursePopup';
  popup.classList.add('popup');

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

  popupHeaderTitle.innerHTML = `<b>${course.Subject} - ${course.Title}</b>`
  popupBody.innerHTML = `${course_to_str(course, session)}`;

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

function displayPopupHandler(course, session) {
  displayPopup(course, session);
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

  popupHeaderTitle.innerHTML = `<b>Schedule ${current_schedule + 1} Weights</b>`

  // Find weights by name
  const gpaWeight = schedule.Weights.find(w => w.Name === 'GPA');
  const startWeight = schedule.Weights.find(w => w.Name === 'Start');
  const endWeight = schedule.Weights.find(w => w.Name === 'End');
  const gapWeight = schedule.Weights.find(w => w.Name === 'GAP');

  // Add content to the popup
  popupBody.innerHTML = `Average Score: ${schedule.Score.toFixed(2)}<br>
                         Average GPA: ${(gpaWeight ? gpaWeight.Value * 4.0 : 'N/A').toFixed(2)}<br>
                         Start Time Score: ${startWeight ? startWeight.Value.toFixed(2) : 'N/A'}<br>
                         End Time Score: ${endWeight ? endWeight.Value.toFixed(2) : 'N/A'}<br>
                         Gaps Score: ${gapWeight ? gapWeight.Value.toFixed(2) : 'N/A'}`;

  // Append the popup parts together and add it to the document
  popupHeader.appendChild(popupHeaderTitle);
  popupHeader.appendChild(popupCloseButton);
  schedulePopup.appendChild(popupHeader);
  schedulePopup.appendChild(popupBody)
  document.body.appendChild(schedulePopup);
}

function clearAsyncCourses() {
  const asyncCoursesList = document.getElementById('async-courses-list');
  asyncCoursesList.innerHTML = '';
  document.getElementById('async-courses-container').style.display = 'none';
}

function displayAsyncCourses() {
  const asyncCoursesContainer = document.getElementById('async-courses-container');
  const asyncCoursesList = document.getElementById('async-courses-list');
  clearAsyncCourses();

  if (!asyncCourses || asyncCourses.length === 0) {
    asyncCoursesContainer.style.display = 'none';
    return;
  }

  asyncCoursesContainer.style.display = 'block';
  asyncCourses.forEach(course => {
    const courseElement = document.createElement('div');
    courseElement.classList.add('async-course-item');
    courseElement.innerHTML = `<strong>${course.Subject} - ${course.Title}</strong><br>CRN: ${course.CRN}`;

    courseElement.addEventListener('click', () => {
      // Assuming the first session is the async one
      displayPopup(course, course.Sessions[0]);
    });

    asyncCoursesList.appendChild(courseElement);
  });
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
    } else {
      // Find the weight with the matching name
      const aWeight = a.Weights.find(w => w.Name.toLowerCase() === sortValue);
      const bWeight = b.Weights.find(w => w.Name.toLowerCase() === sortValue);

      // If both weights are found, compare their values
      if (aWeight && bWeight) {
        return bWeight.Value - aWeight.Value;
      }

      // If a weight is missing, put that schedule last
      if (!aWeight && bWeight) return 1;
      if (aWeight && !bWeight) return -1;

      // If both weights are missing, maintain original order
      return 0;
    }
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
