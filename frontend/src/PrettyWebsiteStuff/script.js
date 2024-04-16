function addTask() {
    var classInput = document.getElementById("classInput");
    var sectionInput = document.getElementById("sectionInput");
    
    var className = classInput.value.trim();
    var section = sectionInput.value;

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
    } else if (className!== ""){
        alert("Please enter a correct number pussy😝");
    }else {
        alert("Please enter a class name 😾");

    }
}
document.addEventListener('DOMContentLoaded', function() {
    fetch('subjects.txt')
      .then(response => response.text())
      .then(data => {
        const lines = data.split('\n');
        const select = document.getElementById('classInput');
        select.innerHTML = '<option value="" selected disabled>Select a class...</option>';
        lines.forEach(function(line) {
          const option = document.createElement('option');
          option.text = line.trim();
          option.value = line.trim();
          select.add(option);
        });
      });
  
    document.getElementById('sectionInput').addEventListener('keypress', function(event) {
      if (event.key === 'Enter') {
        document.getElementById('addClassBtn').click();
      }
    });
  
    document.getElementById('viewClassesBtn').addEventListener('click', function() {
      const selectedClasses = [];
      const select = document.getElementById('classInput');
      const sectionInput = document.getElementById('sectionInput').value;
      const selectedClass = select.options[select.selectedIndex].text;
      if (selectedClass && sectionInput) {
        selectedClasses.push(selectedClass + ' ' + sectionInput);
      }
      // Redirect to the new page with selected classes displayed
      const url = 'displayClasses.html?classes=' + encodeURIComponent(selectedClasses.join(','));
      window.location.href = url;
    });
  });
  

  function displaySelectedClasses() {
    const urlParams = new URLSearchParams(window.location.search);
    const classesParam = urlParams.get('classes');
    const classes = classesParam.split(',');
    const selectedClassesList = document.getElementById('selectedClassesList');
    classes.forEach(function(classInfo) {
      const [className, section] = classInfo.split(' ');
      const li = document.createElement('li');
      li.textContent = `${className} - Section ${section}`;
      selectedClassesList.appendChild(li);
    });
  }
  
  document.addEventListener('DOMContentLoaded', function() {
    displaySelectedClasses();
  });
  