import React, { useState, useEffect } from "react";
import { stringToColor } from "../utils";
import ClassDetailsPopup from "./ClassDetailsPopup";

const ScheduleTable = ({
    schedule,
    currentSchedule,
    totalSchedules,
    onPrevious,
    onNext,
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
        if (schedule && schedule.courses) {
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
        const newTableData = hours.map((hour) => {
            const rowData = { time: hour };
            fullDaysOfWeek.forEach((day) => {
                rowData[day] = { content: null, color: "", fullCourse: null };
            });
            return rowData;
        });

        schedule.courses.forEach((course) => {
            const days = course.days.split("");
            const startTime = parseInt(course.start_time);
            const endTime = parseInt(course.end_time);
            const courseColor = stringToColor(course.crn); // Generate color based on CRN

            days.forEach((day) => {
                const fullDay = fullDaysOfWeek.find((d) => d.startsWith(day));
                for (let time = startTime; time < endTime; time += 100) {
                    const hour = `${Math.floor(time / 100)
                        .toString()
                        .padStart(2, "0")}:00`;
                    const rowIndex = hours.indexOf(hour);
                    if (rowIndex !== -1) {
                        newTableData[rowIndex][fullDay] = {
                            content: {
                                title: course.subject,
                                details: `${course.instructor}\n${course.crn}\n${course.room}`,
                            },
                            color: courseColor,
                            fullCourse: course, // Store the full course object
                        };
                    }
                }
            });
        });

        setTableData(newTableData);
    };

    const hasClasses =
        schedule && schedule.courses && schedule.courses.length > 0;

    const handleCellClick = (course) => {
        setSelectedCourse(course);
    };

    const handleClosePopup = () => {
        setSelectedCourse(null);
    };

    return (
        <div className="w-full">
            <div className="flex justify-center mb-4 space-x-4">
                <button
                    onClick={onPrevious}
                    disabled={!hasClasses || currentSchedule === 0}
                    className="btn btn-primary">
                    Previous Schedule
                </button>
                <button
                    onClick={onNext}
                    disabled={
                        !hasClasses || currentSchedule === totalSchedules - 1
                    }
                    className="btn btn-primary">
                    Next Schedule
                </button>
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
