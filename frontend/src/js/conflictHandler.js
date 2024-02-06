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
      if(course1 === undefined){
          return false;
      }
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
  //TODO: This is close, soomehow is infinit looping
  function finalSchedule(listOfCourses, results = [], currentPermutation = [], nextOptions){
       if (listOfCourses.length === 0) {
          // Add a copy of the current permutation to the results
          results.push([...currentPermutation]);
          return results;
       }
       
       for(let i = 0; i<listOfCourses.length; i++){
           currentPermutation.push(listOfCourses[i]);
           for(let j = i+1; j<listOfCourses.length; j++){
                if(!hasConflict(currentPermutation[-1], listOfCourses[j])){
                    nextOptions.push(listOfCourses[j])
                }
           }
           results.concat(finalSchedule(nextOptions, results, currentPermutation, []));
       }
       return results;
  }

  
  const courses = [
    new Course("Math 1", "MTWRF", "0900", "1030"),
    new Course("Math 2", "TWR", "1100", "1230"),
    new Course("Math 3", "MWF", "1200", "1400"),
    new Course("Math 1", "MTWRF", "0800", "0820")
  ];
  const result = finalSchedule(courses, [], [], []);
