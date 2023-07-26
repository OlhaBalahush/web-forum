const scrollContainer = document.getElementById('category-container');
const content = document.getElementById('btn-category');

const buttonRight = document.getElementById('slideRight');
const buttonLeft = document.getElementById('slideLeft');

buttonRight.onclick = function () {
    document.getElementById('category-container').scrollLeft += 200;
};
buttonLeft.onclick = function () {
    document.getElementById('category-container').scrollLeft -= 200;
};

scrollContainer.addEventListener('wheel', (event) => {
    if (event.deltaY > 0) {
        // Scrolling down
        scrollContainer.scrollTop += 50; // Change the scroll position in the vertical direction
    } else if (event.deltaY < 0) {
        // Scrolling up
        scrollContainer.scrollTop -= 50;
    }

    if (event.deltaX > 0) {
        // Scrolling right
        scrollContainer.scrollLeft += 50; // Change the scroll position in the horizontal direction
    } else if (event.deltaX < 0) {
        // Scrolling left
        scrollContainer.scrollLeft -= 50;
    }
});