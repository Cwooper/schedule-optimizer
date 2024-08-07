import React from "react";

const Footer = () => (
    <footer className="footer bg-lightgray text-white p-10">
        <div className="flex flex-col">
            <h3 className="footer-title">About Us</h3>
            <p>
                This is a schedule optimizer passion project for Western
                Washington University.
            </p>
            <p>
                Created by Cooper Morgan, Konnor Kooi, Robert Bates, Arne
                Wiseman, and Ben Huynh.
            </p>
            <p>
                <a
                    href="https://github.com/cwooper/schedule-optimizer"
                    target="_blank"
                    rel="noopener noreferrer">
                    GitHub Source Code
                </a>
            </p>
        </div>
        <div className="flex flex-col">
            <h3 className="footer-title">Report Issues to</h3>
            <p>
                <a href="mailto:cwooperm@gmail.com">cwooperm@gmail.com</a>
            </p>
        </div>
        <div className="flex flex-col">
            <h3 className="footer-title">Disclaimer</h3>
            <p>
                This project is not affiliated with Western Washington
                University. It is an independent initiative developed solely for
                educational and personal use. All data provided by this project
                is for informational purposes only and should not be considered
                official or binding. Use at your own discretion.
            </p>
        </div>
    </footer>
);

export default Footer;
