document.addEventListener("DOMContentLoaded", function() {
    const orderUidInput = document.getElementById("orderUidInput");
    const submitButton = document.getElementById("submitButton");
    const result = document.getElementById("result");

    submitButton.addEventListener("click", function() {
        const orderUid = orderUidInput.value;

        // Construct the URL with the input order_uid
        const apiUrl = `http://localhost:8080/orders?order_uid=${orderUid}`;

        // Clear previous results
        result.innerHTML = "";

        // Make an HTTP GET request
        fetch(apiUrl)
            .then(response => response.text())
            .then(data => {
                // Display the response on the page
                result.textContent = data;
            })
            .catch(error => {
                // Handle errors
                console.error("Error:", error);
            });
    });
});
