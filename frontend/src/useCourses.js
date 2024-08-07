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

export const useSchedules = () => {
    const [allSchedules, setAllSchedules] = useState([]);
    const [currentSchedule, setCurrentSchedule] = useState(0);
    const [errorMessage, setErrorMessage] = useState("");
    const [warnings, setWarnings] = useState([]); // Add this line

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
        // Create the RawRequest structure
        const rawRequest = {
            Courses: courses.map((course) => course.name),
            Forced: forceList,
            Term: `${term}${quarterCode}`,
            Min: min,
            Max: max,
        };
        try {
            const response = await axios.post(
                "/schedule-optimizer/",
                rawRequest
            );
            const data = response.data;
            console.log(data);
            if (data.Errors && data.Errors.length > 0) {
                setErrorMessage(data.Errors.join("<br>"));
            } else {
                setErrorMessage("");
            }

            if (data.Warnings && data.Warnings.length > 0) {
                setWarnings(data.Warnings); // Set warnings
            } else {
                setWarnings([]);
            }

            if (data.Schedules && data.Schedules.length > 0) {
                setAllSchedules(data.Schedules);
                setCurrentSchedule(0);
            } else {
                setAllSchedules([]);
            }
        } catch (error) {
            console.error("Error:", error);
            setErrorMessage("An error occurred while fetching schedules.");
            setWarnings([]);
        }
    };

    return {
        allSchedules,
        setAllSchedules,
        currentSchedule,
        setCurrentSchedule,
        errorMessage,
        setErrorMessage,
        warnings, // Add this line
        setWarnings, // Add this line
        generateJSON,
    };
};
