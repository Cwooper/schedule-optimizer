import React from "react";
import styles from "./SubmitButton.module.css";

interface SubmitButtonProps {
  onSubmit: () => Promise<void>;
  disabled?: boolean;
}

const SubmitButton: React.FC<SubmitButtonProps> = ({
  onSubmit,
  disabled = false,
}) => {
  const handleClick = async () => {
    try {
      await onSubmit();
    } catch (error) {
      console.error("Error submitting schedule:", error);
      // Throw the error in the message box in course selector
    }
  };

  return (
    <button
      onClick={handleClick}
      disabled={disabled}
      className={styles.submitButton}
      type="button" // Explicitly set type to prevent form submission
    >
      Submit Schedule
    </button>
  );
};

export default SubmitButton;
