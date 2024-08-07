import React, { useState, useCallback, useEffect } from "react";
import Header from "./components/Header";
import CourseForm from "./components/CourseForm";
import CourseList from "./components/CourseList";
import ScheduleTable from "./components/ScheduleTable";
import Footer from "./components/Footer";
import HelpPopup from "./components/HelpPopup";
import { useCourses, useSchedules } from "./useCourses";
import Dropdown from "./components/Dropdown";
import { clearSchedule, addCoursesToCalendar, sortSchedules } from "./utils";

function App() {
    const [min, setMin] = useState(2);
    const [max, setMax] = useState(4);
    const [term, setTerm] = useState("2024");
    const [quarter, setQuarter] = useState("Winter");
    const [globalError, setGlobalError] = useState("");
    const [loading, setLoading] = useState(false);
    const [sortCriteria, setSortCriteria] = useState("score");

    const {
        courses,
        forceList,
        errorMessage,
        setErrorMessage,
        handleAddCourse,
        handleRemoveCourse,
        handleToggleForce,
    } = useCourses();

    const {
        allSchedules,
        setAllSchedules,
        currentSchedule,
        setCurrentSchedule,
        errorMessage: scheduleError,
        setErrorMessage: setScheduleError,
        warnings, // Add this line
        setWarnings, // Add this line
        generateJSON,
    } = useSchedules();

    const handleSort = useCallback(
        (criteria) => {
            setSortCriteria(criteria);
            const sortedSchedules = [...allSchedules].sort((a, b) => {
                if (criteria === "score") {
                    return b.Score - a.Score;
                } else {
                    const weightA = a.Weights.find(
                        (w) => w.Name.toLowerCase() === criteria.toLowerCase()
                    );
                    const weightB = b.Weights.find(
                        (w) => w.Name.toLowerCase() === criteria.toLowerCase()
                    );
                    return weightB.Value - weightA.Value;
                }
            });
            setAllSchedules(sortedSchedules);
            setCurrentSchedule(0);
        },
        [allSchedules, setAllSchedules, setCurrentSchedule]
    );

    const handleNextSchedule = () => {
        if (currentSchedule < allSchedules.length - 1) {
            setCurrentSchedule((prevSchedule) => prevSchedule + 1);
        }
    };

    const handlePreviousSchedule = () => {
        if (currentSchedule > 0) {
            setCurrentSchedule((prevSchedule) => prevSchedule - 1);
        }
    };

    const displaySchedule = useCallback(
        (schedule) => {
            if (!schedule || !schedule.Courses) {
                console.error("Invalid schedule:", schedule);
                return;
            }

            console.log("Displaying schedule:", schedule);
            clearSchedule();
            addCoursesToCalendar(schedule, currentSchedule + 1);

            // Update corner cell with schedule information
            const cornerCell = document.getElementById("cornerCell");
            if (cornerCell) {
                cornerCell.textContent = `Schedule ${currentSchedule + 1} of ${
                    allSchedules.length
                }`;
            }
            console.log(
                `Displaying schedule ${currentSchedule + 1} of ${
                    allSchedules.length
                }`
            );
        },
        [
            currentSchedule,
            allSchedules.length,
            clearSchedule,
            addCoursesToCalendar,
        ]
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

    const handleGenerateJSON = async () => {
        if (validateCourses()) {
            setLoading(true);
            try {
                await generateJSON(courses, forceList, min, max, term, quarter);
            } catch (error) {
                console.error("Error generating schedules:", error);
                setErrorMessage(
                    "An error occurred while generating schedules."
                );
            } finally {
                setLoading(false);
            }
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

    const handleScheduleChange = (index) => {
        if (index < 0 || index >= allSchedules.length) return;
        setCurrentSchedule(index);
    };

    useEffect(() => {
        if (allSchedules.length > 0 && allSchedules[currentSchedule]) {
            displaySchedule(allSchedules[currentSchedule]);
        }
    }, [allSchedules, currentSchedule, displaySchedule]);

    // Add this new useEffect to handle initial sorting
    useEffect(() => {
        if (allSchedules.length > 0) {
            handleSort(sortCriteria);
        }
    }, []);

    return (
        <div className="App bg-white">
            <Header />
            <div className="container mx-auto px-2 py-4 bg-white text-darkgray">
                <div className="flex flex-col md:flex-row gap-4">
                    <div className="bg-white p-2 rounded-lg shadow-md md:w-4/12 w-full flex flex-col items-center">
                        <CourseForm
                            onAddCourse={handleAddCourseWithValidation}
                            courseCount={courses.length}
                        />
                        <div className="w-full pt-2">
                            <CourseList
                                courses={courses}
                                onRemoveCourse={
                                    handleRemoveCourseWithValidation
                                }
                                onToggleForce={handleToggleForceWithValidation}
                            />
                        </div>
                        <button
                            className="mt-2 btn btn-primary bg-darkblue text-white w-1/2"
                            onClick={handleGenerateJSON}
                            disabled={loading}>
                            {loading ? "Generating..." : "Submit"}
                        </button>
                        {errorMessage && (
                            <div
                                id="errorMessage"
                                className="mt-2 text-red-500 text-sm">
                                {errorMessage}
                            </div>
                        )}
                        {warnings.length > 0 && (
                            <div
                                id="warningMessage"
                                className="mt-2 text-red-500 text-sm">
                                {warnings.map((warning, index) => (
                                    <div key={index}>{warning}</div>
                                ))}
                            </div>
                        )}
                        {globalError && (
                            <div
                                id="globalError"
                                className="mt-2 text-red-500 text-sm">
                                {globalError}
                            </div>
                        )}
                    </div>
                    <div className="w-full md:w-3/5">
                        <div className="bg-white p-2 rounded-lg shadow-md flex flex-wrap justify-center justify-evenly">
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
                        <div className="bg-white p-2 rounded-lg shadow-md mt-2">
                            <div>
                                <ScheduleTable
                                    schedule={allSchedules[currentSchedule]}
                                    currentSchedule={currentSchedule}
                                    totalSchedules={allSchedules.length}
                                    onPrevious={handlePreviousSchedule}
                                    onNext={handleNextSchedule}
                                    onSort={handleSort}
                                    sortCriteria={sortCriteria}
                                />
                            </div>
                        </div>
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
