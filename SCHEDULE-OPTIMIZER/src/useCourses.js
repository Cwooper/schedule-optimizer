import { useState, useCallback } from "react";
import { stringToColor, getQuarterCode } from "./utils";

export const useCourses = () => {
    const [courses, setCourses] = useState([]);
    const [forceList, setForceList] = useState([]);
    const [errorMessage, setErrorMessage] = useState("");

    const handleAddCourse = (courseName, section) => {
        if (
            courses.some((course) => course.name === `${courseName} ${section}`)
        ) {
            return false;
        }

        const newCourse = {
            name: `${courseName} ${section}`,
            forced: false,
        };

        setCourses([...courses, newCourse]);
        return true;
    };

    const handleRemoveCourse = (courseName) => {
        setCourses(courses.filter((course) => course.name !== courseName));
        setForceList(forceList.filter((course) => course !== courseName));
    };

    const handleToggleForce = (courseName) => {
        setCourses(
            courses.map((course) =>
                course.name === courseName
                    ? { ...course, forced: !course.forced }
                    : course
            )
        );
        if (forceList.includes(courseName)) {
            setForceList(forceList.filter((course) => course !== courseName));
        } else {
            setForceList([...forceList, courseName]);
        }
    };

    return {
        courses,
        forceList,
        errorMessage,
        setErrorMessage,
        handleAddCourse,
        handleRemoveCourse,
        handleToggleForce,
    };
};

export const useSchedules = (
    setAllSchedules,
    setCurrentSchedule,
    displayScheduleFn, // Renamed to avoid conflict
    currentSchedule,
    setErrorMessage, // Added setErrorMessage parameter
    sortSchedules // Added sortSchedules parameter
) => {
    const generateJSON = (courses, forceList, min, max, term, quarter) => {
        if (courses.length < min) {
            setErrorMessage(
                "Cannot have less courses than minimum courses in a schedule."
            );
            return;
        }

        if (forceList.length > max) {
            setErrorMessage(
                "Cannot have more forced courses than maximum courses in a schedule."
            );
            return;
        }

        const quarterCode = getQuarterCode(quarter);

        const scheduleinfo = {
            courses: courses.map((course) => course.name),
            force: forceList,
            min: min,
            max: max,
            term: `${term}${quarterCode}`,
        };

        const json = JSON.stringify(scheduleinfo);
        console.log(json);

        fetch("/schedule-optimizer", {
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
                    setErrorMessage(response.errors.join("<br>"));
                    if (response.warnings && response.warnings.length > 0) {
                        setErrorMessage(
                            (prev) =>
                                prev + "<br>" + response.warnings.join("<br>")
                        );
                    }
                } else {
                    if (response.warnings && response.warnings.length > 0) {
                        setErrorMessage(response.warnings.join("<br>"));
                    }
                    setAllSchedules(response.schedules);
                    sortSchedules(response.schedules);
                    setCurrentSchedule(0);
                    displayScheduleFn(response.schedules[0]);
                }
            })
            .catch((error) => {
                console.error("Error:", error);
            });
    };

    const addCoursesToCalendar = useCallback(
        (schedule) => {
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
        },
        [currentSchedule]
    );

    const clearSchedule = useCallback(() => {
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
    }, []);

    const displaySchedule = useCallback(
        (schedule) => {
            if (!schedule || !schedule.courses) {
                console.error("Invalid schedule:", schedule);
                return;
            }

            clearSchedule();
            addCoursesToCalendar(schedule);
        },
        [clearSchedule, addCoursesToCalendar]
    );

    return {
        generateJSON,
        addCoursesToCalendar,
        clearSchedule,
        displaySchedule,
    };
};
