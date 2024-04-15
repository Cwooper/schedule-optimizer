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

function removeTask(button) {
    var task = button.parentNode;
    task.parentNode.removeChild(task);
}
