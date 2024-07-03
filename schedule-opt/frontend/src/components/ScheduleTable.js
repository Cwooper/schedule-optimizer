import React, { useEffect } from "react";

const ScheduleTable = ({
    schedule,
    currentSchedule,
    displaySchedule,
    clearSchedule,
}) => {
    useEffect(() => {
        createTable();
        displaySchedule(schedule[currentSchedule]);
    }, [schedule, currentSchedule, displaySchedule]);

    const createTable = () => {
        const hours = [
            "0800",
            "0900",
            "1000",
            "1100",
            "1200",
            "1300",
            "1400",
            "1500",
            "1600",
            "1700",
        ];
        const fullDaysOfWeek = [
            "Monday",
            "Tuesday",
            "Wednesday",
            "Thursday",
            "Friday",
        ];
        let table = document.getElementById("calendar");
        if (table) {
            table.innerHTML = ""; // Clear existing table
        } else {
            table = document.createElement("table");
            table.id = "calendar";
            document.body.appendChild(table);
        }

        const headerRow = document.createElement("tr");
        const cornerCell = document.createElement("th");
        cornerCell.id = "cornerCell";
        headerRow.appendChild(cornerCell);

        fullDaysOfWeek.forEach((day) => {
            const headerCell = document.createElement("th");
            headerCell.textContent = day;
            headerRow.appendChild(headerCell);
        });

        table.appendChild(headerRow);

        hours.forEach((hour) => {
            const row = document.createElement("tr");
            const hourCell = document.createElement("td");
            hourCell.textContent = `${hour.slice(0, 2)}:${hour.slice(2)}`;
            row.appendChild(hourCell);

            fullDaysOfWeek.forEach((day) => {
                const cell = document.createElement("td");
                cell.id = `${day}-${hour}`;
                row.appendChild(cell);
            });

            table.appendChild(row);
            table.classList.add("calendar-table");
        });
    };

    return <div id="calendar"></div>;
};

export default ScheduleTable;
