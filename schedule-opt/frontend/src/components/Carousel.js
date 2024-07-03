import React from "react";
import useEmblaCarousel from "embla-carousel-react";
import "./Carousel.css";

export const Carousel = ({ children, className }) => {
    const [emblaRef] = useEmblaCarousel();
    return (
        <div className={`embla ${className}`} ref={emblaRef}>
            <div className="embla__container">{children}</div>
        </div>
    );
};

export const CarouselContent = ({ children }) => (
    <div className="embla__slide">{children}</div>
);

export const CarouselItem = ({ children }) => (
    <div className="embla__slide__inner">{children}</div>
);

export const CarouselNext = ({ onClick }) => (
    <button className="embla__button embla__button--next" onClick={onClick}>
        Next
    </button>
);

export const CarouselPrevious = ({ onClick }) => (
    <button className="embla__button embla__button--prev" onClick={onClick}>
        Previous
    </button>
);
