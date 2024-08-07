// utils.js
export const stringToColor = (str) => {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const hue = hash % 360;
    return `hsla(${hue}, 70%, 80%, 0.95)`; // Lighter shade with 60% opacity
};

export const getQuarterCode = (quarter) => {
    switch (quarter) {
        case "Winter":
            return 10;
        case "Spring":
            return 20;
        case "Fall":
            return 30;
        case "Summer":
            return 40;
        default:
            return 0;
    }
};

export const sortSchedules = (schedules, sortValue) => {
    if (schedules.length <= 0) return;

    schedules.sort((a, b) => {
        if (sortValue === "score") {
            return b.score - a.score;
        } else if (sortValue === "end") {
            return b.weights.end - a.weights.end;
        } else if (sortValue === "gap") {
            return b.weights.gap - a.weights.gap;
        } else if (sortValue === "gpa") {
            return b.weights.gpa - a.weights.gpa;
        } else if (sortValue === "start") {
            return b.weights.start - a.weights.start;
        }
        return 0;
    });
};

export const clearSchedule = () => {
    const table = document.getElementById("calendar-table");
    if (!table) {
        // console.error("No element with ID 'calendar-table' found.");
        return;
    }
    const cells = table.getElementsByTagName("td");
    for (let cell of cells) {
        if (cell.id !== "cornerCell" && !cell.classList.contains("time-cell")) {
            cell.innerHTML = "";
            cell.style.backgroundColor = "";
            cell.classList.remove("scheduled-course");
        }
    }
};

export const addCoursesToCalendar = (schedule, currentSchedule) => {
    console.log("Adding courses to calendar:", schedule, currentSchedule);
    const courses = schedule.Courses;
    const cornerCell = document.getElementById("cornerCell");
    if (cornerCell) {
        cornerCell.textContent = `Schedule ${currentSchedule}`;
    }

    courses.forEach((course) => {
        course.Sessions.forEach((session, index) => {
            const days = session.Days ? session.Days.split("") : [];
            const startTime = parseInt(session.StartTime);
            const endTime = parseInt(session.EndTime);
            const isLab = index > 0;

            days.forEach((day) => {
                let dayLetter;
                switch (day) {
                    case "M":
                        dayLetter = "M";
                        break;
                    case "T":
                        dayLetter = "T";
                        break;
                    case "W":
                        dayLetter = "W";
                        break;
                    case "R":
                        dayLetter = "R";
                        break; // Thursday
                    case "F":
                        dayLetter = "F";
                        break;
                    default:
                        return; // Skip if it's not a valid day
                }

                let startHour = Math.floor(startTime / 100);
                let endHour = Math.ceil(endTime / 100);
                for (let i = startHour; i < endHour; i++) {
                    const hour = i < 10 ? `0${i}00` : `${i}00`;
                    const cellId = `${dayLetter}-${hour}`;
                    const cell = document.getElementById(cellId);
                    if (cell) {
                        console.log(`Populating cell ${cellId}`);
                        const bgColor = stringToColor(course.CRN.toString());
                        cell.style.backgroundColor = bgColor;
                        cell.innerHTML = `<div class="text-xs p-1">
                <b>${isLab ? `${course.Subject} LAB` : course.Subject}</b><br>
                ${session.Instructor}<br>
                ${course.CRN}<br>
                ${session.Location}
              </div>`;
                        cell.classList.add("scheduled-course");
                    }
                }
            });
        });
    });
};

// export const addCoursesToCalendar = (schedule, currentSchedule) => {
//     console.log("Adding courses to calendar:", schedule, currentSchedule);
//     const courses = schedule.courses;
//     const cornerCell = document.getElementById("cornerCell");
//     if (cornerCell) {
//         cornerCell.textContent = `Schedule ${currentSchedule}`;
//     }

//     courses.forEach((course) => {
//         // console.log("Processing course:", course);
//         const days = course.days ? course.days.split("") : [];
//         const startTime = parseInt(course.start_time);
//         const endTime = parseInt(course.end_time);

//         days.forEach((day) => {
//             let startHour = Math.floor(startTime / 100);
//             let endHour = Math.ceil(endTime / 100);
//             for (let i = startHour; i < endHour; i++) {
//                 const hour = i < 10 ? `0${i}00` : `${i}00`;
//                 const cellId = `${day}-${hour}`;
//                 const cell = document.getElementById(cellId);
//                 if (cell) {
//                     console.log(`Populating cell ${cellId}`);
//                     const bgColor = stringToColor(course.crn);
//                     cell.style.backgroundColor = bgColor;
//                     cell.innerHTML = `<div class="text-xs p-1">
//                         <b>${course.subject}</b><br>
//                         ${course.instructor}<br>
//                         ${course.crn}<br>
//                         ${course.room}
//                     </div>`;
//                     cell.classList.add("scheduled-course");
//                 } // else {
//                 //     console.log(`Cell ${cellId} not found`);
//                 // }
//             }
//         });

//         if (course.lab_days) {
//             const labDays = course.lab_days ? course.lab_days.split("") : [];
//             const labStart = parseInt(course.lab_start_time);
//             const labEnd = parseInt(course.lab_end_time);
//             labDays.forEach((labDay) => {
//                 let labStartHour = Math.floor(labStart / 100);
//                 let labEndHour = Math.ceil(labEnd / 100);
//                 for (let i = labStartHour; i < labEndHour; i++) {
//                     if (i < 10) {
//                         i = `0${i}`;
//                     }
//                     const cellId = `${labDay}-${i}00`;
//                     const cell = document.getElementById(cellId);

//                     if (cell) {
//                         const bgColor = stringToColor(course.crn);
//                         cell.style.backgroundColor = bgColor;
//                         cell.innerHTML = `<b>${course.subject} LAB</b><br>${course.instructor}<br>${course.crn}<br>${course.lab_room}`;
//                         cell.classList.add("scheduled-course");
//                     }
//                 }
//             });
//         }
//     });
// };
