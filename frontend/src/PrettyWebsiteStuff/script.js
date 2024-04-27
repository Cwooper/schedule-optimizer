function addTask() {
  var classInput = document.getElementById("classInput");
  var sectionInput = document.getElementById("sectionInput");

  var className = classInput.value.trim();
  var section = sectionInput.value.trim();

  if (className !== "" && section !== "") {
    var taskList = document.getElementById("taskList");
    var task = document.createElement("div");
    task.className = "task";
    task.innerHTML = `
      <div class="class-item">
          <div>
              <span>Class: ${className}</span><br>
              <span>Section: ${section}</span>
          </div>
          <button onclick="removeTask(this)">Remove</button>
      </div>
    `;
    taskList.appendChild(task);
    classInput.value = "";
    sectionInput.value = "";
    classInput.selectedIndex = 0; // Reset dropdown to default option
  } else if (className !== "") {
    alert("Please enter a section number.");
  } else {
    alert("Please enter a class name.");
  }
}
function displaySelectedClasses() {
  const urlParams = new URLSearchParams(window.location.search);
  const classesParam = urlParams.get('classes');
  const classes = classesParam.split(',');
  const selectedClassesList = document.getElementById('selectedClassesList');
  classes.forEach(function (classInfo) {
    const [className, section] = classInfo.split(' ');
    const li = document.createElement('li');
    li.textContent = `${className} - Section ${section}`;
    selectedClassesList.appendChild(li);
  });

  // Store selected classes in local storage
  localStorage.setItem('selectedClasses', JSON.stringify(classes));

  // Add button to go back to main page
  const backButton = document.createElement('button');
  backButton.textContent = 'Back to Main Page';
  backButton.addEventListener('click', function () {
    // Redirect to main page and pass selected classes from local storage
    const selectedClasses = localStorage.getItem('selectedClasses');
    const url = 'index.html?classes=' + encodeURIComponent(selectedClasses);
    window.location.href = url;
  });
  document.body.appendChild(backButton);
}

document.addEventListener('DOMContentLoaded', function () {
  displaySelectedClasses();
});


document.addEventListener('DOMContentLoaded', function() {
  fetch('subjects.txt')
    .then(response => response.text())
    .then(data => {
      const lines = data.split('\n');
      const select = document.getElementById('classInput');
      select.innerHTML = '<option value="" selected disabled>Select a class...</option>';
      lines.forEach(function (line) {
        const option = document.createElement('option');
        option.text = line.trim();
        option.value = line.trim();
        select.add(option);
      });
    });

  document.getElementById('sectionInput').addEventListener('keypress', function (event) {
    if (event.key === 'Enter') {
      document.getElementById('addClassBtn').click();
    }
  });

  document.getElementById('viewClassesBtn').addEventListener('click', function () {
    const selectedClasses = [];
    const select = document.getElementById('classInput');
    const sectionInput = document.getElementById('sectionInput').value.trim();
    const selectedClass = select.options[select.selectedIndex].text;
    if (selectedClass && sectionInput) {
      selectedClasses.push(selectedClass + ' ' + sectionInput);
    }
    // Redirect to the new page with selected classes displayed
    const url = 'displayClasses.html?classes=' + encodeURIComponent(selectedClasses.join(','));
    window.location.href = url;
  });

  function displaySelectedClasses() {
    const urlParams = new URLSearchParams(window.location.search);
    const classesParam = urlParams.get('classes');
    const classes = classesParam.split(',');
    const selectedClassesList = document.getElementById('selectedClassesList');
    classes.forEach(function (classInfo) {
      const [className, section] = classInfo.split(' ');
      const li = document.createElement('li');
      li.textContent = `${className} - Section ${section}`;
      selectedClassesList.appendChild(li);
    });
  }

  displaySelectedClasses();
});
