import React from "react";

function Header() {
    const displayHelpPopup = () => {
        document.getElementById("help-popup").classList.add("block");
        document.getElementById("help-popup").classList.remove("hidden");
        document.getElementById("backdrop").classList.add("block");
        document.getElementById("backdrop").classList.remove("hidden");
    };

    return (
        <header className="bg-white border-b border-gray-200">
            <nav className="max-w-7xl mx-auto px-2 sm:px-6 lg:px-8">
                <div className="relative flex items-center justify-between h-16">
                    <div className="flex-1 flex items-center justify-between">
                        <span className="text-2xl font-bold text-darkblue">
                            WWU Schedule Optimizer
                        </span>
                        <button
                            className="btn btn-primary bg-darkblue text-white"
                            onClick={displayHelpPopup}
                            title="Display Help Menu">
                            Help
                        </button>
                    </div>
                </div>
            </nav>
        </header>
    );
}

export default Header;
