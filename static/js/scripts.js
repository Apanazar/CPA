document.addEventListener("DOMContentLoaded", function () {
  console.log("The DOM is fully loaded and disassembled");

  const uploadBtn = document.getElementById("uploadBtn");
  const pixlInput = document.getElementById("pixlInput");
  const downloadBtn = document.getElementById("downloadBtn");
  const pixelitcanvas = document.getElementById("pixelitcanvas");
  const blocksize = document.getElementById("blocksize");
  const blockvalue = document.getElementById("blockvalue");
  const paletteselector = document.getElementById("paletteselector");
  const useAllColors = document.getElementById("useAllColors");
  const useAllColorsValue = document.getElementById("useAllColorsValue");
  const brightness = document.getElementById("brightness");
  const brightnessValue = document.getElementById("brightnessValue");
  const contrast = document.getElementById("contrast");
  const contrastValue = document.getElementById("contrastValue");
  const saturation = document.getElementById("saturation");
  const saturationValue = document.getElementById("saturationValue");
  const contour = document.getElementById("contour");
  const resolutionSelect = document.getElementById("resolution"); 

  let originalImage = null;
  let palettesInfo = {};

  new SlimSelect({
    select: "#paletteselector",
    placeholder: "Select a palette",
  });

  fetch("/palettes")
    .then((response) => {
      if (!response.ok) {
        throw new Error("Palettes could not be loaded");
      }
      return response.json();
    })
    .then((data) => {
      console.log("Palettes have been received:", data);
      data.forEach((paletteInfo) => {
        const option = document.createElement("option");
        option.value = paletteInfo.name;
        option.textContent = capitalizeFirstLetter(
          paletteInfo.name.replace("-", " ")
        );
        option.dataset.count = paletteInfo.count;
        paletteselector.appendChild(option);
        palettesInfo[paletteInfo.name] = paletteInfo.count;
      });

      const firstOption = paletteselector.options[0];
      if (firstOption) {
        const paletteCount = parseInt(firstOption.dataset.count, 10);
        if (paletteCount && !isNaN(paletteCount)) {
          useAllColors.max = paletteCount;
          useAllColors.value = 1;
          useAllColorsValue.textContent = "1";
        }
      }
    })
    .catch((error) => {
      console.error("Error loading palettes:", error);
    });

  uploadBtn.addEventListener("click", () => {
    console.log("The download button is pressed");
    pixlInput.click();
  });

  downloadBtn.addEventListener("click", () => {
    const canvas = document.getElementById("pixelitcanvas");
    if (canvas) {
      const image = canvas.toDataURL("image/png");
      const link = document.createElement("a");
      link.href = image;
      link.download = "processed-image.png";
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      console.log("Image download initiated");
    } else {
      alert("There is no image to download!");
    }
  });

  pixlInput.addEventListener("change", (event) => {
    const file = event.target.files[0];
    if (file && file.type.startsWith("image/")) {
      const reader = new FileReader();
      reader.onload = function (e) {
        console.log("Image uploaded");
        originalImage = e.target.result;
        processImage();
        downloadBtn.disabled = false; 
      };
      reader.readAsDataURL(file);
    } else {
      alert("Please select a valid image file..");
      console.log("The selected file is not an image");
    }
  });

  blocksize.addEventListener("input", () => {
    blockvalue.textContent = blocksize.value;
    console.log("The block size has been changed:", blocksize.value);
    processImage();
  });

  useAllColors.addEventListener("input", () => {
    useAllColorsValue.textContent = useAllColors.value;
    console.log(
      "The degree of use of all colors of the palette has been changed:",
      useAllColors.value
    );
    processImage();
  });

  brightness.addEventListener("input", () => {
    brightnessValue.textContent = brightness.value;
    console.log("The brightness has been changed:", brightness.value);
    processImage();
  });

  contrast.addEventListener("input", () => {
    contrastValue.textContent = contrast.value;
    console.log("The contrast has been changed:", contrast.value);
    processImage();
  });

  saturation.addEventListener("input", () => {
    saturationValue.textContent = saturation.value;
    console.log("The saturation has been changed:", saturation.value);
    processImage();
  });

  contour.addEventListener("change", () => {
    console.log("Contour:", contour.checked);
    processImage();
  });

  paletteselector.addEventListener("change", () => {
    console.log("The palette is selected:", paletteselector.value);
    const selectedOption =
      paletteselector.options[paletteselector.selectedIndex];
    const paletteCount = parseInt(selectedOption.dataset.count, 10);
    if (paletteCount && !isNaN(paletteCount)) {
      useAllColors.max = paletteCount;
      if (parseInt(useAllColors.value, 10) > paletteCount) {
        useAllColors.value = paletteCount;
        useAllColorsValue.textContent = "1";
      }
      console.log(
        `Maximum number of colors for useAllColors: ${paletteCount}`
      );
    } else {
      useAllColors.max = 20;
      useAllColors.value = 1;
      useAllColorsValue.textContent = "1";
    }
    processImage();
  });
  resolutionSelect.addEventListener("change", () => {
    console.log("Resolution selected:", resolutionSelect.value);
    processImage();
  });
 
  function processImage() {
    if (!originalImage) {
      console.log("There is no source image to process");
      return;
    }
    console.log("Starting image processing");

    const img = new Image();
    img.crossOrigin = "Anonymous";
    img.onload = function () {
      const tempCanvas = document.createElement("canvas");
      tempCanvas.width = img.width;
      tempCanvas.height = img.height;
      const ctx = tempCanvas.getContext("2d");
      ctx.drawImage(img, 0, 0);
 
      uploadToServer(tempCanvas);
    };
    img.onerror = function () {
      console.error("Couldn't upload the original image.");
    };
    img.src = originalImage;
  }

  function uploadToServer(canvas) {
    canvas.toBlob(function (blob) {
      if (!blob) {
        console.error("Couldn't create Blob from image.");
        return;
      }

      const formData = new FormData();
      formData.append("image", blob, "upload.png");
      formData.append("blocksize", blocksize.value);
      formData.append("palette", paletteselector.value);
      formData.append("useAllColors", useAllColors.value);
      formData.append("brightness", brightness.value);
      formData.append("contrast", contrast.value);
      formData.append("saturation", saturation.value);
      formData.append("contour", contour.checked ? "on" : "off");
      formData.append("resolution", resolutionSelect.value);

      console.log("Sending data to the server");

      fetch("/process", {
        method: "POST",
        body: formData,
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error("Error in image processing.");
          }
          return response.blob();
        })
        .then((blob) => {
          const url = URL.createObjectURL(blob);
          const ctx = pixelitcanvas.getContext("2d");
          const processedImg = new Image();
          processedImg.onload = function () {
            pixelitcanvas.width = processedImg.width;
            pixelitcanvas.height = processedImg.height;
            ctx.drawImage(processedImg, 0, 0);
            URL.revokeObjectURL(url);
            console.log("The processed image is rendered on canvas");
          };
          processedImg.src = url;
        })
        .catch((error) => {
          console.error("Error in image processing:", error);
        });
    }, "image/png");
  }

  function capitalizeFirstLetter(string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
  }
});