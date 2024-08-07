import React, { useState, useEffect } from "react";
import Dropdown from "./Dropdown";

const CourseForm = ({ onAddCourse, courseCount }) => {
    const [courseName, setCourseName] = useState("");
    const [section, setSection] = useState("");
    const [subjects, setSubjects] = useState([]);
    const [error, setError] = useState("");

    useEffect(() => {
        fetch("/subjects.txt")
            .then((response) => response.text())
            .then((text) => {
                const courses = text.trim().split("\n");
                setSubjects(courses);
                if (!courseName) {
                    setCourseName(courses[0]); // Set the default value to the first subject
                }
            })
            .catch((error) => {
                console.error("Error fetching subjects:", error);
            });
    }, [courseName]);

    const handleAddCourse = () => {
        if (!section) {
            setError("Course number is required");
            return;
        }
        if (courseCount >= 13) {
            setError("Cannot add more than 13 courses");
            return;
        }
        setError("");
        if (courseName && onAddCourse(courseName, section)) {
            setSection(""); // Only reset the section field
        }
    };

    const handleSectionChange = (e) => {
        const value = e.target.value;
        if (/^\d*$/.test(value)) {
            setSection(value);
            if (value && error === "Course number is required") {
                setError(""); // Clear error if section is provided
            }
        }
    };

    const handleKeyPress = (e) => {
        if (e.key === "Enter") {
            e.preventDefault(); // Prevent form submission
            handleAddCourse();
        }
    };

    return (
        <div className="px-4 flex flex-col items-center w-full">
            <div className="flex flex-wrap gap-4 w-full items-end">
                <div className="w-full md:w-auto flex-grow">
                    <Dropdown
                        label="Select Course"
                        value={courseName}
                        options={subjects}
                        onChange={(e) => {
                            setCourseName(e.target.value);
                            if (
                                error === "Cannot add more than 13 courses" &&
                                courseCount < 13
                            ) {
                                setError(""); // Clear error if course count is below 13
                            }
                        }}
                        className="w-full"
                    />
                </div>
                <div className="w-full md:w-auto flex-grow">
                    <div className="w-full">
                        <input
                            type="text"
                            value={section}
                            placeholder="Course Number"
                            onChange={handleSectionChange}
                            onKeyPress={handleKeyPress}
                            className="input input-bordered w-full bg-white border-lightgray"
                        />
                    </div>
                </div>
            </div>
            <button
                className="btn btn-primary mt-4 bg-darkblue text-white w-1/2"
                onClick={handleAddCourse}>
                Add Course
            </button>
            {error && <div className="mt-2 text-red-500">{error}</div>}
        </div>
    );
};

export default CourseForm;
