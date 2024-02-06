// conflictHandler.js
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
    console.log('------------------------');
  }
}

// Returns a list of all nonconflicting schedules
function possibleSchedules(listOfCourses) {
  if (listOfCourses.length === 0) {                    // Base Case
    return;
  }
  const permutations = allPermutations(listOfCourses); // List of List of Courses
  console.log("Permutations\n");
  console.log(permutations);
  const conflicts = allConflicts(listOfCourses);       // Pairs of conflicts
  console.log("Conflicts\n");
  console.log(conflicts);

  // Remove conflicts from permutations
  const results = removeConflicts(permutations, conflicts);
  console.log("Results\n");
  console.log(results);
  return results;
}

// Returns true if course1 and course2 have a time conflict
function hasConflict(course1, course2) {
  for(let i = 0; i < course1.days.length; i++){
    const day = course1.days.charAt(i);
    if(course2.days.includes(day)){
      if((((course1.timeStart>=course2.timeStart) && (course1.timeStart<=course2.timeEnd)) || 
         ((course1.timeEnd>=course2.timeStart) && (course1.timeEnd<=course2.timeEnd))) ||
          (course1.className === course2.className)){
        return true;
      }
    }
  }
  return false;
}

// Returns all non-duplicate permutations of the listOfCourses
function allPermutations(listOfCourses, results = [], currentPermutation = [], start = 0) {
  if (start === listOfCourses.length) {
    // Add a copy of the current permutation to the results
    results.push([...currentPermutation]);
    return results;
  }

  for (let i = start; i < listOfCourses.length; i++) {
    currentPermutation.push(listOfCourses[i]);

    // Recursively generate permutations for the remaining courses
    allPermutations(listOfCourses, results, currentPermutation, i + 1);

    // Backtrack: Remove the last course to explore other permutations
    currentPermutation.pop();
  }

  return results;
} 

// Finds all Conflicts between each pair of courses
function allConflicts(listOfCourses) {
  let conflicts = [];
  for (let i = 0; i < listOfCourses.length; i++) {
    for (let j = i + 1; j < listOfCourses.length; j++) {
      if (hasConflict(listOfCourses[i], listOfCourses[j])) {
        conflicts.push([listOfCourses[i], listOfCourses[j]]);
      }
    }
  }
  return conflicts;
}

// Returns permutations without the conflicting pairs 
function removeConflicts(permutations, conflicts) {
  let results = permutations;
  for (const conflict of conflicts) {
    for (let i = 0; i < permutations.length; i++) {
      if (permutations[i].includes(conflict[0]) &&
          permutations[i].includes(conflict[1])) {
        results.splice(i, 1);
        i--;
      }
    }
  }
  return results;
}

const courses = [
  new Course("Math 1", "MTWRF", "0900", "1030"),
  new Course("Math 2", "TWR", "1100", "1230"),
  new Course("Math 3", "MWF", "1200", "1400"),
  new Course("Math 1", "MTWRF", "0800", "0820")
];
const result = possibleSchedules(courses);
