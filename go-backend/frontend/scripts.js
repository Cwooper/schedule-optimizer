// Get references to HTML elements
const courseSelect = document.getElementById("courseSelect");
const sectionNumber = document.getElementById("sectionNumber");
const courseList = document.getElementById("courseList");

let courses = [];
let forceList = [];
let allSchedules = [];
let currentSchedule = 0;

function addCourse() {
  const courseName = courseSelect.value;
  const section = sectionNumber.value;
  const errorMessage = document.getElementById("errorMessage");
  errorMessage.textContent = "";

  if (
    courses.some(
      (course) =>
        course.Subject === courseName && course.CRN === parseInt(section)
    )
  ) {
    errorMessage.textContent = "Duplicate Course Entered.";
    return;
  }

  if (!sectionNumber.checkValidity()) {
    errorMessage.textContent = "Please enter a valid section number.";
    return;
  }

  if (courseName && section && courses.length < 10) {
    const newCourse = courseName + " " + section;
    courses.push(newCourse);
    const li = document.createElement("li");
    courseList.appendChild(li);

    const forceButton = document.createElement("button");
    forceButton.textContent = "Force";
    forceButton.addEventListener("click", function () {
      const index = forceList.findIndex(
        (c) => c.Subject === newCourse.Subject && c.CRN === newCourse.CRN
      );
      if (index !== -1) {
        forceList.splice(index, 1);
        forceButton.classList.remove("forced");
      } else {
        forceList.push(newCourse);
        forceButton.classList.add("forced");
      }
    });

    const removeButton = document.createElement("button");
    removeButton.textContent = "–";
    removeButton.addEventListener("click", function () {
      errorMessage.textContent = "";
      const index = courses.findIndex(
        (c) => c.Subject === newCourse.Subject && c.CRN === newCourse.CRN
      );
      if (index !== -1) {
        courses.splice(index, 1);
      }
      const forceIndex = forceList.findIndex(
        (c) => c.Subject === newCourse.Subject && c.CRN === newCourse.CRN
      );
      if (forceIndex !== -1) {
        forceList.splice(forceIndex, 1);
      }
      li.remove();
      if (courses.length === 0) {
        submitButton.style.display = "none";
      }
    });

    const textContainer = document.createElement("div");
    textContainer.textContent = `${courseName} ${section}`;
    textContainer.style.flex = "1";
    li.appendChild(textContainer);
    li.appendChild(forceButton);
    li.appendChild(removeButton);
    li.style.display = "flex";
    li.style.justifyContent = "space-between";

    if (courses.length > 0) {
      submitButton.style.display = "block";
    }

    errorMessage.textContent = "";
  } else {
    errorMessage.textContent = "Ensure you have less than 10 classes added.";
  }
}

function generateJSON() {
  const min = minSelect.value;
  const max = maxSelect.value;
  const term = termSelect.value;
  const quarter = quarterSelect.value;
  const errorMessage = document.getElementById("errorMessage");
  errorMessage.textContent = "";

  if (courses.length < min) {
    errorMessage.textContent +=
      "Cannot have less courses than minimum courses in a schedule.";
    return;
  }

  if (forceList.length > max) {
    errorMessage.textContent +=
      "Cannot have more forced courses than maximum courses in a schedule.";
    return;
  }

  const scheduleInfo = {
    Courses: courses, // string[]
    Forced: forceList, // string[]
    Term: term + quarter, // string
    Min: parseInt(min), // int
    Max: parseInt(max), // int
  };

  const json = JSON.stringify(scheduleInfo);
  console.log(json);

  fetch("/schedule-optimizer/", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: json,
  })
    .then((response) => response.json())
    .then((response) => {
      console.log(response);
      if (response.errors && response.errors.length > 0) {
        errorMessage.innerHTML = response.errors.join("<br>");
        if (response.warnings && response.warnings.length > 0) {
          errorMessage.innerHTML += "<br>" + response.warnings.join("<br>");
        }
      } else {
        if (response.warnings && response.warnings.length > 0) {
          errorMessage.innerHTML = response.warnings.join("<br>");
        }
        allSchedules = response.Schedules;
        sortSchedules();
        currentSchedule = 0;
        displaySchedule(allSchedules[currentSchedule]);
      }
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}

function createTable() {
  const hours = [
    "0800",
    "0900",
    "1000",
    "1100",
    "1200",
    "1300",
    "1400",
    "1500",
    "1600",
    "1700",
  ];
  const daysOfWeek = ["M", "T", "W", "R", "F"];
  const fullDaysOfWeek = [
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
  ];
  const table = document.getElementById("calendar");

  const headerRow = document.createElement("tr");
  const cornerCell = document.createElement("th");
  cornerCell.id = "cornerCell";
  headerRow.appendChild(cornerCell);

  fullDaysOfWeek.forEach((day) => {
    const headerCell = document.createElement("th");
    headerCell.textContent = day;
    headerRow.appendChild(headerCell);
  });

  table.appendChild(headerRow);

  hours.forEach((hour) => {
    const row = document.createElement("tr");
    const hourCell = document.createElement("td");
    hourCell.textContent = `${hour.slice(0, 2)}:${hour.slice(2)}`;
    row.appendChild(hourCell);

    daysOfWeek.forEach((day) => {
      const cell = document.createElement("td");
      cell.id = `${day}-${hour}`;
      row.appendChild(cell);
    });

    table.appendChild(row);
    table.classList.add("calendar-table");
  });
}

function stringToColor(str) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  const c = (hash & 0x00ffffff).toString(16).toUpperCase().padStart(6, "0");
  return `#${c}`;
}

function addCoursesToCalendar(schedule) {
  const cornerCell = document.getElementById("cornerCell");
  cornerCell.textContent = `Schedule ${currentSchedule}`;
  cornerCell.addEventListener("mouseover", () =>
    addHoverEffect.call(cornerCell, "lightgray")
  );
  cornerCell.addEventListener("mouseout", () =>
    removeHoverEffect.call(cornerCell, "lightgray")
  );
  cornerCell.addEventListener("click", () =>
    displaySchedulePopupHandler(schedule)
  );

  schedule.Courses.forEach((course) => {
    course.Sessions.forEach((session) => {
      if (!session.IsAsync && !session.IsTimeTBD) {
        const days = session.Days.split("");
        const startTime = parseInt(session.StartTime);
        const endTime = parseInt(session.EndTime);

        days.forEach((day) => {
          let startHour = Math.floor(startTime / 100);
          let endHour = Math.ceil(endTime / 100);
          for (let i = startHour; i < endHour; i++) {
            const cellId = `${day}-${i.toString().padStart(2, "0")}00`;
            const cell = document.getElementById(cellId);
            const bgColor = stringToColor(course.CRN.toString());

            cell.addEventListener("mouseover", () =>
              addHoverEffect.call(cell, bgColor)
            );
            cell.addEventListener("mouseout", () =>
              removeHoverEffect.call(cell, bgColor)
            );
            cell.addEventListener("click", () => displayPopupHandler(course));

            cell.style.backgroundColor = bgColor;
            cell.innerHTML = `<b>${course.Subject}</b><br>${session.Instructor}<br>${course.CRN}<br>${session.Location}`;
            cell.classList.add("scheduled-course");
          }
        });
      }
    });
  });
}

function displaySchedule(schedule) {
  clearSchedule();
  addCoursesToCalendar(schedule);
}

function clearSchedule() {
  const table = document.getElementById("calendar");
  const elements = table.getElementsByTagName("*");

  for (let i = 0; i < elements.length; i++) {
    const element = elements[i];
    if (element.id) {
      element.textContent = "";
      element.style = "";
      element.classList.remove("scheduled-course");
      element.removeEventListener("mouseover", element._mouseover);
      element.removeEventListener("mouseout", element._mouseout);
      element.removeEventListener("click", element._click);
    }
  }
}

function nextSchedule() {
  if (currentSchedule + 1 < allSchedules.length) {
    currentSchedule++;
    errorMessage.textContent = "";
    displaySchedule(allSchedules[currentSchedule]);
  } else {
    errorMessage.textContent = "You are at the end of the schedules";
  }
}

function prevSchedule() {
  if (currentSchedule - 1 > -1) {
    currentSchedule--;
    errorMessage.textContent = "";
    displaySchedule(allSchedules[currentSchedule]);
  } else {
    errorMessage.textContent = "You are at the beginning of the schedules";
  }
}

function course_to_str(course) {
  let result = `Instructor: ${course.Sessions[0].Instructor}<br>
    CRN: ${course.CRN}<br>`;

  result += course.GPA ? `Average GPA: ${course.GPA}<br>` : "";

  result += `Credits: ${course.Credits}<br>
    Room: ${course.Sessions[0].Location}<br>`;

  result += course.Sessions[0].Days
    ? `Days: ${course.Sessions[0].Days}<br>
    Times: ${course.Sessions[0].StartTime} - ${course.Sessions[0].EndTime}<br>`
    : "";

  result += `Available Seats: ${course.AvailableSeats}<br>
    Max Students: ${course.Capacity}<br>
    Students Enrolled: ${course.Enrolled}<br>`;

  result += course.WaitlistCount ? `Waitlist: ${course.WaitlistCount}<br>` : "";
  result += course.Prerequisites
    ? `Prerequisites: ${course.Prerequisites}<br>`
    : "";
  result += course.Attributes ? `Attributes: ${course.Attributes}<br>` : "";
  result += course.AdditionalFees
    ? `Additional Fees: ${course.AdditionalFees}`
    : "";

  if (result.endsWith("<br>")) {
    result = result.slice(0, -4);
  }

  return result;
}

function displayPopup(course) {
  const oldPopup = document.getElementById("coursePopup");
  if (oldPopup) {
    oldPopup.remove();
  }
  const popup = document.createElement("div");
  popup.id = "coursePopup";
  popup.classList.add("popup");

  const popupHeader = document.createElement("div");
  popupHeader.classList.add("popup-header");
  const popupHeaderTitle = document.createElement("div");
  popupHeaderTitle.classList.add("popup-header-title");
  const popupCloseButton = document.createElement("button");
  popupCloseButton.classList.add("popup-close-button");
  popupCloseButton.innerHTML = "&times;";
  popupCloseButton.onclick = () => {
    if (popup) {
      popup.remove();
    }
  };

  const popupBody = document.createElement("div");
  popupBody.classList.add("popup-body");

  popupHeaderTitle.innerHTML = `<b>${course.Subject}</b>`;
  popupBody.innerHTML = course_to_str(course);

  popupHeader.appendChild(popupHeaderTitle);
  popupHeader.appendChild(popupCloseButton);
  popup.appendChild(popupHeader);
  popup.appendChild(popupBody);
  document.body.appendChild(popup);
}

function handleKeyPress(event) {
  if (event.keyCode === 13) {
    addCourse();
    document.getElementById("sectionNumber").value = "";
  }
}

function addHoverEffect(bgColor) {
  this.style.backgroundColor = "lightblue";
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
  const oldPopup = document.getElementById("schedulePopup");
  if (oldPopup) {
    oldPopup.remove();
  }
  const schedulePopup = document.createElement("div");
  schedulePopup.id = "schedulePopup";
  schedulePopup.classList.add("popup");

  const popupHeader = document.createElement("div");
  popupHeader.classList.add("popup-header");
  const popupHeaderTitle = document.createElement("div");
  popupHeaderTitle.classList.add("popup-header-title");
  const popupCloseButton = document.createElement("button");
  popupCloseButton.classList.add("popup-close-button");
  popupCloseButton.innerHTML = "&times;";
  popupCloseButton.onclick = () => {
    if (schedulePopup) {
      schedulePopup.remove();
    }
  };

  const popupBody = document.createElement("div");
  popupBody.classList.add("popup-body");

  popupHeaderTitle.innerHTML = `<b>Schedule ${currentSchedule} Weights</b>`;
  popupBody.innerHTML = `Average Score: ${schedule.Score}<br>
                           Average GPA: ${
                             schedule.Weights.find((w) => w.Name === "GPA")
                               .Value * 4.0
                           }<br>
                           Start Time Score: ${
                             schedule.Weights.find((w) => w.Name === "Start")
                               .Value
                           }<br>
                           End Time Score: ${
                             schedule.Weights.find((w) => w.Name === "End")
                               .Value
                           }<br>
                           Gaps Score: ${
                             schedule.Weights.find((w) => w.Name === "GAP")
                               .Value
                           }`;

  popupHeader.appendChild(popupHeaderTitle);
  popupHeader.appendChild(popupCloseButton);
  schedulePopup.appendChild(popupHeader);
  schedulePopup.appendChild(popupBody);
  document.body.appendChild(schedulePopup);
}

function sortSchedules() {
  if (allSchedules.length <= 0) {
    return;
  }

  const sortButton = document.getElementById("scheduleSort");
  const sortValue = sortButton.value;

  allSchedules.sort((a, b) => {
    if (sortValue === "score") {
      return b.Score - a.Score;
    } else if (sortValue === "nd") {
      return (
        b.Weights.find((w) => w.Name === "End").Value -
        a.Weights.find((w) => w.Name === "End").Value
      );
    } else if (sortValue === "gap") {
      return (
        b.Weights.find((w) => w.Name === "GAP").Value -
        a.Weights.find((w) => w.Name === "GAP").Value
      );
    } else if (sortValue === "gpa") {
      return (
        b.Weights.find((w) => w.Name === "GPA").Value -
        a.Weights.find((w) => w.Name === "GPA").Value
      );
    } else if (sortValue === "start") {
      return (
        b.Weights.find((w) => w.Name === "Start").Value -
        a.Weights.find((w) => w.Name === "Start").Value
      );
    }
    return 0;
  });

  console.log(allSchedules)
}

function updateSort() {
  if (allSchedules.length <= 0) {
    return;
  }
  sortSchedules();
  currentSchedule = 0;
  displaySchedule(allSchedules[currentSchedule]);
}

// Fetch the class names from subjects.txt and populate the dropdown menu
fetch("subjects.txt")
  .then((response) => response.text())
  .then((text) => {
    const courses = text.trim().split("\n");
    courses.forEach((courseName) => {
      const option = document.createElement("option");
      option.value = courseName;
      option.textContent = courseName;
      courseSelect.appendChild(option);
    });
  })
  .catch((error) => {
    console.error("Error fetching subjects:", error);
  });

function displayHelpPopup() {
  const helpPopup = document.getElementById("help-popup");
  helpPopup.classList.add("active");
}

function closeHelpPopup() {
  const helpPopup = document.getElementById("help-popup");
  helpPopup.classList.remove("active");
}

createTable();

// Event listeners
document.getElementById("addCourseButton").addEventListener("click", addCourse);
document
  .getElementById("generateButton")
  .addEventListener("click", generateJSON);
document.getElementById("nextButton").addEventListener("click", nextSchedule);
document.getElementById("prevButton").addEventListener("click", prevSchedule);
document.getElementById("scheduleSort").addEventListener("change", updateSort);
document
  .getElementById("helpButton")
  .addEventListener("click", displayHelpPopup);
document
  .getElementById("closeHelpButton")
  .addEventListener("click", closeHelpPopup);
