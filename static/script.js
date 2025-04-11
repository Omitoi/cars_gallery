// Function to get cookie by name
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

// Function to set cookie
function setCookie(name, value, days = 365) {
    const date = new Date();
    date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
    const expires = `expires=${date.toUTCString()}`;
    document.cookie = `${name}=${value};${expires};path=/`;
}

// Function to toggle favorite
function toggleFavorite(carId) {
    const btn = document.querySelector(`.favorite-btn[data-car-id="${carId}"]`);
    if (!btn) return;

    const favorites = getCookie('favorites') ? getCookie('favorites').split(',').filter(Boolean) : [];
    
    // Get car name from either regular car card, mini car card, or single car view
    let carName = '';
    const carCard = btn.closest('.car-card, .mini-car-card');
    if (carCard) {
        carName = carCard.querySelector('label').textContent;
    } else {
        // In single car view
        const columns = btn.closest('.content-single').querySelector('.columns');
        if (columns) {
            carName = columns.querySelector('h3:nth-child(2)').textContent;
        }
    }
    
    if (favorites.includes(carId.toString())) {
        // Remove from favorites
        const index = favorites.indexOf(carId.toString());
        favorites.splice(index, 1);
        btn.classList.remove('active');
        btn.setAttribute('aria-label', `Add ${carName} to favorites`);
    } else {
        // Add to favorites at the beginning of the array to maintain newest first order
        favorites.unshift(carId.toString());
        btn.classList.add('active');
        btn.setAttribute('aria-label', `Remove ${carName} from favorites`);
    }
    
    setCookie('favorites', favorites.join(','));
}

// Initialize favorite buttons on page load
document.addEventListener('DOMContentLoaded', function() {
    const favorites = getCookie('favorites') ? getCookie('favorites').split(',') : [];
    favorites.forEach(carId => {
        const btn = document.querySelector(`.favorite-btn[data-car-id="${carId}"]`);
        if (btn) {
            btn.classList.add('active');
        }
    });
});

// Image dragging functionality for comparison page
document.addEventListener("DOMContentLoaded", function () {
    const image = document.querySelector(".content-single .image-box img");
    const imageBox = document.querySelector(".content-single .image-box");

    if (!image) return;
    image.draggable = false;
    let isDragging = false;
    let startY, posY = 0;
    let limit = Math.max(0, (image.clientHeight - imageBox.clientHeight) / 2);
    image.style.cursor = "grab";

    image.addEventListener("mousedown", function (e) {
        isDragging = true;
        startY = e.clientY;
        image.style.cursor = "grabbing";
    });

    document.addEventListener("mousemove", function (e) {
        if (!isDragging) return;

        let deltaY = e.clientY - startY;
        let final = posY + deltaY;

        // Constrain movement within limits
        final = Math.min(limit, Math.max(-limit, final));

        image.style.top = `${final}px`;
    });

    document.addEventListener("mouseup", function () {
        if (isDragging) {
            let computedStyle = getComputedStyle(image);
            posY = parseFloat(computedStyle.top);
            isDragging = false;
        }
        image.style.cursor = "grab";
    });
}); 