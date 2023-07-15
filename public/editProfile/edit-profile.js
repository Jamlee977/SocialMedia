const emailInput = document.getElementById('email');
const firstNameInput = document.getElementById('first_name');
const lastNameInput = document.getElementById('last_name');
const confirmButton = document.getElementById('confirm_button');
const backButton = document.getElementById('back_button');

const email = emailInput.value;
const firstName = firstNameInput.value;
const lastName = lastNameInput.value;

emailInput.addEventListener("input", () => {
    if (emailInput.value !== email && confirmButton.disabled) {
        confirmButton.disabled = false;
    }
});

firstNameInput.addEventListener("input", () => {
    if (firstNameInput.value !== firstName && confirmButton.disabled) {
        confirmButton.disabled = false;
    }
});

lastNameInput.addEventListener("input", () => {
    if (lastNameInput.value !== lastName && confirmButton.disabled) {
        confirmButton.disabled = false;
    }
});
