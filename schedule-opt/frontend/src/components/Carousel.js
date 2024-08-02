import React, { useCallback, useEffect } from "react";
import useEmblaCarousel from "embla-carousel-react";
import "./Carousel.css"; // Ensure this path is correct

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

export const Carousel = ({ children, className, onSlideChange }) => {
    const [emblaRef, emblaApi] = useEmblaCarousel({ loop: false });

    const scrollPrev = useCallback(() => {
        if (emblaApi) emblaApi.scrollPrev();
    }, [emblaApi]);

    const scrollNext = useCallback(() => {
        if (emblaApi) emblaApi.scrollNext();
    }, [emblaApi]);

    useEffect(() => {
        if (emblaApi) {
            emblaApi.on("select", () => {
                onSlideChange(emblaApi.selectedScrollSnap());
            });
        }
    }, [emblaApi, onSlideChange]);

    return (
        <div className={`embla ${className}`}>
            <div className="embla__viewport" ref={emblaRef}>
                <div className="embla__container">{children}</div>
            </div>
            <CarouselPrevious onClick={scrollPrev} />
            <CarouselNext onClick={scrollNext} />
        </div>
    );
};

export const CarouselContent = ({ children }) => (
    <div className="embla__slide">{children}</div>
);

export const CarouselItem = ({ children }) => (
    <div className="embla__slide__inner">{children}</div>
);
