const shortenForm = document.getElementById('shorten-form');
const longUrlInput = document.getElementById('long-url');
const customCodeInput = document.getElementById('custom-code');
const resultDiv = document.getElementById('result');

shortenForm.addEventListener('submit', (event) => {
    event.preventDefault();

    const longUrl = longUrlInput.value;
    const customCode = customCodeInput.value;

    fetch('/api/url/shorten', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json' 
        },
        body: JSON.stringify({
            url: longUrl,
            ...(!!customCode && { short_hash: customCode })
        })
    })
    .then(response => response.json())
    .then(res => {
        if (res.error) {
            resultDiv.textContent = res.error
            return
        }

        const link = document.createElement('a');
        link.href = `/api/url/redirect/${res?.short_hash}`;
        link.target = '_blank';
        link.textContent = res?.short_hash;

        resultDiv.textContent = 'Your URL hash: ';
        resultDiv.appendChild(link); 
    })
    .catch(error => {
        console.error('Error:', error);
        resultDiv.textContent = 'An error occurred while shortening the URL.'; 
    });
});
