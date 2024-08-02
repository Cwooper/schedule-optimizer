import { useState, useCallback } from "react";
import { stringToColor, getQuarterCode } from "./utils";
import axios from "axios";

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
    displayScheduleFn,
    setErrorMessage,
    sortSchedules
) => {
    const generateJSON = async (
        courses,
        forceList,
        min,
        max,
        term,
        quarter
    ) => {
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

        try {
            const response = await axios.post(
                "/api/schedule-optimizer",
                scheduleinfo
            );
            const data = response.data;
            console.log(data);
            if (data.errors && data.errors.length > 0) {
                setErrorMessage(data.errors.join("<br>"));
                if (data.warnings && data.warnings.length > 0) {
                    setErrorMessage(
                        (prev) => prev + "<br>" + data.warnings.join("<br>")
                    );
                }
            } else {
                if (data.warnings && data.warnings.length > 0) {
                    setErrorMessage(data.warnings.join("<br>"));
                }
                setAllSchedules(data.schedules);
                sortSchedules(data.schedules);
                setCurrentSchedule(0);
            }
        } catch (error) {
            console.error("Error:", error);
            setErrorMessage("An error occurred while fetching schedules.");
        }
    };

    return {
        generateJSON,
    };
};
