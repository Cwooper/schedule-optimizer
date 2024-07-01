import React, { useState, useCallback } from "react";
import Header from "./components/Header";
import CourseForm from "./components/CourseForm";
import CourseList from "./components/CourseList";
import ScheduleTable from "./components/ScheduleTable";
import Footer from "./components/Footer";
import HelpPopup from "./components/HelpPopup";
import { useCourses, useSchedules } from "./useCourses";
import {
    stringToColor,
    getQuarterCode,
    sortSchedules,
    clearSchedule,
    addCoursesToCalendar,
} from "./utils";
import Dropdown from "./components/Dropdown";

function App() {
    const [allSchedules, setAllSchedules] = useState([]);
    const [currentSchedule, setCurrentSchedule] = useState(0);
    const [min, setMin] = useState(2);
    const [max, setMax] = useState(4);
    const [term, setTerm] = useState("2024");
    const [quarter, setQuarter] = useState("Winter");
    const [globalError, setGlobalError] = useState("");

    const {
        courses,
        forceList,
        errorMessage,
        setErrorMessage,
        handleAddCourse,
        handleRemoveCourse,
        handleToggleForce,
    } = useCourses();

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

    const { generateJSON } = useSchedules(
        setAllSchedules,
        setCurrentSchedule,
        displaySchedule,
        currentSchedule,
        setErrorMessage,
        sortSchedules
    );

    const validateCourses = () => {
        if (courses.length === 0) {
            setGlobalError("You must select at least one course");
            return false;
        }
        if (courses.length > 10) {
            setGlobalError("You cannot select more than 10 courses");
            return false;
        }
        if (courses.length < min) {
            setGlobalError(
                `Cannot have less courses than minimum courses in a schedule. Minimum courses: ${min}`
            );
            return false;
        }
        setGlobalError("");
        return true;
    };

    const handleGenerateJSON = () => {
        if (validateCourses()) {
            generateJSON(courses, forceList, min, max, term, quarter);
        }
    };

    const handleAddCourseWithValidation = (courseName, section) => {
        if (globalError) setGlobalError("");
        return handleAddCourse(courseName, section);
    };

    const handleRemoveCourseWithValidation = (courseName) => {
        if (globalError) setGlobalError("");
        handleRemoveCourse(courseName);
    };

    const handleToggleForceWithValidation = (courseName) => {
        if (globalError) setGlobalError("");
        handleToggleForce(courseName);
    };

    return (
        <div className="App bg-white">
            <Header />
            <div className="container mx-auto px-4 py-6 bg-white text-darkgray">
                <div className="flex flex-col md:flex-row gap-6">
                    <div className="bg-white p-4 rounded-lg shadow-md md:w-4/12 w-full flex flex-col items-center">
                        <CourseForm
                            onAddCourse={handleAddCourseWithValidation}
                            courseCount={courses.length}
                        />
                        <div className="w-full pt-4">
                            <CourseList
                                courses={courses}
                                onRemoveCourse={
                                    handleRemoveCourseWithValidation
                                }
                                onToggleForce={handleToggleForceWithValidation}
                            />
                        </div>
                        <button
                            className="mt-4 btn btn-primary bg-darkblue text-white w-1/2"
                            onClick={handleGenerateJSON}>
                            Submit
                        </button>
                        {errorMessage && (
                            <div
                                id="errorMessage"
                                className="mt-4 text-red-500">
                                {errorMessage}
                            </div>
                        )}
                        {globalError && (
                            <div id="globalError" className="mt-4 text-red-500">
                                {globalError}
                            </div>
                        )}
                    </div>
                    <div className="w-full md:w-3/5">
                        <div className="bg-white p-4 rounded-lg shadow-md flex flex-wrap  justify-center justify-evenly">
                            <Dropdown
                                label="Minimum Courses:"
                                value={min}
                                options={[...Array(6).keys()].map((i) => i + 1)}
                                onChange={(e) => {
                                    setMin(e.target.value);
                                    if (globalError) setGlobalError(""); // Reset error when min is changed
                                }}
                                className="w-full md:w-1/4"
                            />
                            <Dropdown
                                label="Maximum Courses:"
                                value={max}
                                options={[...Array(10).keys()].map(
                                    (i) => i + 1
                                )}
                                onChange={(e) => {
                                    setMax(e.target.value);
                                    if (globalError) setGlobalError(""); // Reset error when max is changed
                                }}
                                className="w-full md:w-1/4"
                            />
                            <Dropdown
                                label="Year:"
                                value={term}
                                options={[2024, 2025]}
                                onChange={(e) => {
                                    setTerm(e.target.value);
                                    if (globalError) setGlobalError(""); // Reset error when term is changed
                                }}
                                className="w-full md:w-1/4"
                            />
                            <Dropdown
                                label="Quarter:"
                                value={quarter}
                                options={["Winter", "Spring", "Summer", "Fall"]}
                                onChange={(e) => {
                                    setQuarter(e.target.value);
                                    if (globalError) setGlobalError(""); // Reset error when quarter is changed
                                }}
                                className="w-full md:w-1/4"
                            />
                        </div>
                        <ScheduleTable
                            schedule={allSchedules}
                            currentSchedule={currentSchedule}
                            displaySchedule={displaySchedule}
                            clearSchedule={clearSchedule}
                            className="w-full mt-4"
                        />
                    </div>
                </div>
            </div>
            <Footer />
            <div
                id="backdrop"
                className="hidden fixed inset-0 bg-gray-600 bg-opacity-50 z-40"></div>
            <div
                id="help-popup"
                className="hidden fixed inset-0 flex justify-center items-center z-50">
                <HelpPopup />
            </div>
        </div>
    );
}

export default App;
