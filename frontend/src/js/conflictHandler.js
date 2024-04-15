
class Course {
	constructor(className, days, timeStart, timeEnd) {
		this.className = className;
		this.days = days;
		this.timeStart = timeStart;
		this.timeEnd = timeEnd;
	}

	displayInfo() {
		console.log(`Class Name: ${this.className}`);
		console.log(`Days: ${this.days}`);
		console.log(`Time Start: ${this.timeStart}`);
		console.log(`Time End: ${this.timeEnd}`);
	}
}

function findSchedules(listOfCourses, conflicts, currentSchedule = [], step = 0, result = []) {
	result.push(currentSchedule.slice());
	console.log(step + ": " + currentSchedule);
	
	for (let i = step; i < listOfCourses.length; i++) {
		const course = listOfCourses[i];
		// TODO Check if adding the current course conflicts with any/the last course added in the schedule
		// Add the current course to the current schedule
		currentSchedule.push(course);
		
		// Recur with the updated schedule and the next step
		findSchedules(listOfCourses, conflicts, currentSchedule, i + 1, result);
		
		// Backtrack: remove the last course added to try other possibilities
		currentSchedule.pop();
	}
	
	return result;
}

// Finds all Conflicts between each pair of courses
function allConflicts(listOfCourses) {
    let conflicts = new Object();
    for (let i = 0; i < listOfCourses.length; i++) {
        for (let j = i + 1; j < listOfCourses.length; j++) {
            if (hasConflict(listOfCourses[i], listOfCourses[j])) {
                if (!conflicts.hasOwnProperty(listOfCourses[i])) {
                    conflicts[listOfCourses[i]] = new Set();
                }
                if (!conflicts.hasOwnProperty(listOfCourses[j])) {
                    conflicts[listOfCourses[j]] = new Set();
                }
                conflicts[listOfCourses[i]].add(listOfCourses[j]);
                conflicts[listOfCourses[j]].add(listOfCourses[i]);
            }
        }
    }
    return conflicts;
}


// Returns true if course1 and course2 have a time conflict
function hasConflict(course1, course2) {
	for (let i = 0; i < course1.days.length; i++){
		const day = course1.days.charAt(i);
		if (course2.days.includes(day)){
			if ((((course1.timeStart>=course2.timeStart) && (course1.timeStart<=course2.timeEnd)) || 
				((course1.timeEnd>=course2.timeStart) && (course1.timeEnd<=course2.timeEnd))) ||
				(course1.className === course2.className)) {
				return true;
			}
		}
	}
	return false;
}

const courses = [
	new Course("Math 1", "MTWRF", "0900", "1030"),
	new Course("Math 2", "TR", "1100", "1230"),
	new Course("Math 3", "MWF", "1200", "1400"),
	new Course("Math 4", "MTWRF", "0800", "0820"),
	new Course("English 1", "MW", "1000", "1130"),
	new Course("English 2", "TWF", "0900", "1030"),
];

console.log("-----Courses-----");
console.log(courses)

conflicts = allConflicts(courses);

console.log("-----Conflicts-----");
console.log(conflicts)

console.log("-----Finding Schedules-----");
result = findSchedules(courses, conflicts);

console.log("-----Schedules-----");
console.log(result)