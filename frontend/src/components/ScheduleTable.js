import React, { useState, useEffect } from "react";
import { stringToColor } from "../utils";
import ClassDetailsPopup from "./ClassDetailsPopup";
import Dropdown from "./Dropdown";

const ScheduleTable = ({
    schedule,
    currentSchedule,
    totalSchedules,
    onPrevious,
    onNext,
    onSort,
    sortCriteria,
}) => {
    const [tableData, setTableData] = useState([]);
    const [selectedCourse, setSelectedCourse] = useState(null);
    const hours = [
        "08:00",
        "09:00",
        "10:00",
        "11:00",
        "12:00",
        "13:00",
        "14:00",
        "15:00",
        "16:00",
        "17:00",
    ];
    const fullDaysOfWeek = [
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
    ];

    useEffect(() => {
        if (schedule && schedule.Courses) {
            updateScheduleDisplay(schedule);
        } else {
            setTableData(createEmptyTableData());
        }
    }, [schedule, currentSchedule]);

    const createEmptyTableData = () => {
        return hours.map((hour) => {
            const rowData = { time: hour };
            fullDaysOfWeek.forEach((day) => {
                rowData[day] = { content: null, color: "" };
            });
            return rowData;
        });
    };

    const updateScheduleDisplay = (schedule) => {
        const newTableData = createEmptyTableData();

        schedule.Courses.forEach((course) => {
            course.Sessions.forEach((session, index) => {
                const days = session.Days.split("");
                const startTime = parseInt(session.StartTime);
                const endTime = parseInt(session.EndTime);
                const courseColor = stringToColor(course.CRN.toString());

                const isLab = index > 0; // Treat additional sessions as labs

                days.forEach((day) => {
                    let fullDay;
                    switch (day) {
                        case "M":
                            fullDay = "Monday";
                            break;
                        case "T":
                            fullDay = "Tuesday";
                            break;
                        case "W":
                            fullDay = "Wednesday";
                            break;
                        case "R":
                            fullDay = "Thursday";
                            break;
                        case "F":
                            fullDay = "Friday";
                            break;
                        default:
                            return; // Skip if it's not a valid day
                    }

                    for (let time = startTime; time < endTime; time += 100) {
                        const hour = `${Math.floor(time / 100)
                            .toString()
                            .padStart(2, "0")}:00`;
                        const rowIndex = hours.indexOf(hour);
                        if (rowIndex !== -1) {
                            newTableData[rowIndex][fullDay] = {
                                content: {
                                    title: isLab
                                        ? `${course.Subject} LAB`
                                        : course.Subject,
                                    details: `${session.Instructor}\n${course.CRN}\n${session.Location}`,
                                },
                                color: courseColor,
                                fullCourse: course,
                            };
                        }
                    }
                });
            });
        });

        setTableData(newTableData);
    };

    const hasClasses =
        schedule && schedule.Courses && schedule.Courses.length > 0;

    const handleCellClick = (course) => {
        setSelectedCourse(course);
    };

    const handleClosePopup = () => {
        setSelectedCourse(null);
    };

    return (
        <div className="w-full">
            <div className="flex flex-col sm:flex-row items-center justify-between sm:justify-center mb-4 space-y-4 sm:space-y-0 sm:space-x-4 w-full">
                <div className="w-full sm:w-auto order-2 sm:order-1">
                    <button
                        onClick={onPrevious}
                        disabled={currentSchedule === 0}
                        className="btn btn-primary w-full sm:w-auto">
                        Previous Schedule
                    </button>
                </div>
                <div className="w-full sm:w-1/3 flex justify-center order-1 sm:order-2 pb-4 sm:pb-0">
                    <Dropdown
                        label="Sort by"
                        value={sortCriteria}
                        options={[
                            { value: "score", label: "Score" },
                            { value: "end", label: "End Time" },
                            { value: "gap", label: "Gap Time" },
                            { value: "gpa", label: "GPA" },
                            { value: "start", label: "Start Time" },
                        ]}
                        onChange={(e) => onSort(e.target.value)}
                        className="w-40"
                    />
                </div>
                <div className="w-full sm:w-auto order-3">
                    <button
                        onClick={onNext}
                        disabled={currentSchedule === totalSchedules - 1}
                        className="btn btn-primary w-full sm:w-auto">
                        Next Schedule
                    </button>
                </div>
            </div>
            <div className="w-full overflow-x-auto relative">
                {!hasClasses && (
                    <div className="absolute inset-0 bg-gray-200 bg-opacity-75 flex items-center justify-center z-10">
                        <p className="text-xl font-bold text-gray-700">
                            Please add classes to view the schedules
                        </p>
                    </div>
                )}
                <table className="w-full border-collapse text-xs sm:text-base">
                    <thead>
                        <tr className="bg-gray-200">
                            <th className="border border-gray-300 p-1 sm:p-2 text-center font-bold">
                                {hasClasses
                                    ? `Schedule ${
                                          currentSchedule + 1
                                      } of ${totalSchedules}`
                                    : "Time"}
                            </th>
                            {fullDaysOfWeek.map((day) => (
                                <th
                                    key={day}
                                    className="border border-gray-300 p-1 sm:p-2 text-center font-bold">
                                    {day}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {tableData.map((row, index) => (
                            <tr key={index}>
                                <td className="border border-gray-300 p-1 sm:p-2 text-center font-bold">
                                    {row.time}
                                </td>
                                {fullDaysOfWeek.map((day) => (
                                    <td
                                        key={day}
                                        className={`border border-gray-300 p-1 sm:p-2 text-center ${
                                            row[day].content
                                                ? "cursor-pointer hover:opacity-80"
                                                : ""
                                        }`}
                                        style={{
                                            backgroundColor: row[day].color,
                                        }}
                                        onClick={() =>
                                            row[day].content &&
                                            handleCellClick(row[day].fullCourse)
                                        }>
                                        {row[day].content && (
                                            <>
                                                <div className="font-bold">
                                                    {row[day].content.title}
                                                </div>
                                                <div className="text-xs">
                                                    {row[day].content.details}
                                                </div>
                                            </>
                                        )}
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            {selectedCourse && (
                <ClassDetailsPopup
                    course={selectedCourse}
                    onClose={handleClosePopup}
                />
            )}
        </div>
    );
};

export default ScheduleTable;
