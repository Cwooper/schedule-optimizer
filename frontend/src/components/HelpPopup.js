import React from "react";

function HelpPopup() {
    const closeHelpPopup = () => {
        document.getElementById("help-popup").classList.add("hidden");
        document.getElementById("help-popup").classList.remove("block");
        document.getElementById("backdrop").classList.add("hidden");
        document.getElementById("backdrop").classList.remove("block");
    };

    return (
        <div
            id="help-popup"
            className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border lg:w-3/12 sm:w-8/12 shadow-lg rounded-md bg-white">
                <div className="mt-3 text-center">
                    <h3 className="text-2xl leading-6 font-medium text-gray-900 pb-3 border-b border-gray-300">
                        Help Menu
                    </h3>
                    <button
                        onClick={closeHelpPopup}
                        className="absolute top-0 right-0 mt-2 mr-4 text-gray-400 hover:text-gray-600 text-3xl">
                        &times;
                    </button>
                    <div className="px-7 py-3 text-left">
                        <p className="text-sm text-gray-600">
                            To add courses, select a subject, and enter a
                            section number.
                        </p>
                        <ul className="list-disc pl-5 mt-2 text-sm text-gray-600">
                            <li>
                                <b>Minimum Courses:</b> The maximum number of
                                courses that are included per schedule.
                            </li>
                            <li>
                                <b>Maximum Courses:</b> The minimum number of
                                courses that are included per schedule.
                            </li>
                            <li>
                                <b>Year:</b> The year that you want to search
                                for.
                            </li>
                            <li>
                                <b>Quarter:</b> The quarter you want to search
                                for.
                            </li>
                            <li>
                                <b>Force:</b> Courses that are forced (required)
                                to be included in every schedule.
                            </li>
                        </ul>
                        <p className="mt-4 text-sm text-gray-600">
                            You can view more details about a course by clicking
                            it. <br />
                            Also, view more details about a schedule such as
                            Average GPA by pressing the "Schedule Number" in the
                            top left corner!
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default HelpPopup;
