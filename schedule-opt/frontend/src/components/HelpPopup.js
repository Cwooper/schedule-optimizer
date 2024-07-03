import React from "react";

function HelpPopup() {
    const closeHelpPopup = () => {
        document.getElementById("help-popup").classList.add("hidden");
        document.getElementById("help-popup").classList.remove("block");
        document.getElementById("backdrop").classList.add("hidden");
        document.getElementById("backdrop").classList.remove("block");
    };

    return (
        <div className="help-popup p-4 bg-white rounded-lg shadow-lg max-w-2xl mx-auto mt-20 relative z-50">
            <div className="popup-header flex justify-between items-center">
                <div className="popup-header-title text-xl font-bold text-lightgray">
                    Help Menu
                </div>
                <button
                    className="btn btn-sm btn-error"
                    onClick={closeHelpPopup}>
                    &times;
                </button>
            </div>
            <div className="popup-body mt-4 text-lightgray">
                <p>
                    To add courses, select a subject, and enter a section
                    number.
                </p>
                <ul className="list-disc pl-5 mt-2">
                    <li>
                        <b>Minimum Courses:</b> The maximum number of courses
                        that are included per schedule.
                    </li>
                    <li>
                        <b>Maximum Courses:</b> The minimum number of courses
                        that are included per schedule.
                    </li>
                    <li>
                        <b>Year:</b> The year that you want to search for.
                    </li>
                    <li>
                        <b>Quarter:</b> The quarter you want to search for.
                    </li>
                    <li>
                        <b>Force:</b> Courses that are forced (required) to be
                        included in every schedule.
                    </li>
                </ul>
                <p className="mt-4">
                    You can view more details about a course by clicking it.{" "}
                    <br />
                    Also, view more details about a schedule such as Average GPA
                    by pressing the "Schedule Number" in the top left Corner!
                </p>
            </div>
        </div>
    );
}

export default HelpPopup;
