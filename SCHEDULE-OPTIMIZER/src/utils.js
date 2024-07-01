// utils.js

export const stringToColor = (str) => {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const c = (hash & 0x00ffffff).toString(16).toUpperCase().padStart(6, "0");
    return `#${c}`;
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
    const table = document.getElementById("calendar");
    const elements = table.getElementsByTagName("*");

    for (let i = 0; i < elements.length; i++) {
        const element = elements[i];
        if (element.id) {
            element.textContent = "";
            element.style = "";
            element.classList.remove("scheduled-course");
        }
    }
};

export const addCoursesToCalendar = (
    schedule,
    currentSchedule,
    stringToColor
) => {
    const courses = schedule.courses;
    const cornerCell = document.getElementById("cornerCell");
    cornerCell.textContent = `Schedule ${currentSchedule}`;

    courses.forEach((course) => {
        const days = course.days.split("");
        const startTime = parseInt(course.start_time);
        const endTime = parseInt(course.end_time);

        days.forEach((day) => {
            let startHour = Math.floor(startTime / 100);
            let endHour = Math.ceil(endTime / 100);
            for (let i = startHour; i < endHour; i++) {
                if (i < 10) {
                    i = `0${i}`;
                }
                const cellId = `${day}-${i}00`;
                const cell = document.getElementById(cellId);
                const bgColor = stringToColor(course.crn);

                cell.style.backgroundColor = bgColor;
                cell.innerHTML = `<b>${course.subject}</b><br>${course.instructor}<br>${course.crn}<br>${course.room}`;
                cell.classList.add("scheduled-course");
            }
        });

        if (course.lab_days) {
            const labDays = course.lab_days.split("");
            const labStart = parseInt(course.lab_start_time);
            const labEnd = parseInt(course.lab_end_time);
            labDays.forEach((labDay) => {
                let labStartHour = Math.floor(labStart / 100);
                let labEndHour = Math.ceil(labEnd / 100);
                for (let i = labStartHour; i < labEndHour; i++) {
                    if (i < 10) {
                        i = `0${i}`;
                    }
                    const cellId = `${labDay}-${i}00`;
                    const cell = document.getElementById(cellId);
                    const bgColor = stringToColor(course.crn);

                    cell.style.backgroundColor = bgColor;
                    cell.innerHTML = `<b>${course.subject} LAB</b><br>${course.instructor}<br>${course.crn}<br>${course.lab_room}`;
                    cell.classList.add("scheduled-course");
                }
            });
        }
    });
};
