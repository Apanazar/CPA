
document.addEventListener("DOMContentLoaded", function () {
  console.log("DOM полностью загружен и разобран");

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

  // <-- Новый код для разрешения
  const resolutionSelect = document.getElementById("resolution"); 
  // Конец нового кода

  let originalImage = null;
  let palettesInfo = {};

  // Инициализация SlimSelect для палитры
  new SlimSelect({
    select: "#paletteselector",
    placeholder: "Выберите палитру",
  });

  // Загрузка палитр с сервера
  fetch("/palettes")
    .then((response) => {
      if (!response.ok) {
        throw new Error("Не удалось загрузить палитры");
      }
      return response.json();
    })
    .then((data) => {
      console.log("Палитры получены:", data);
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
      console.error("Ошибка при загрузке палитр:", error);
    });

  uploadBtn.addEventListener("click", () => {
    console.log("Кнопка загрузки нажата");
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
      console.log("Скачивание изображения инициировано");
    } else {
      alert("Нет изображения для скачивания!");
    }
  });

  pixlInput.addEventListener("change", (event) => {
    const file = event.target.files[0];
    if (file && file.type.startsWith("image/")) {
      const reader = new FileReader();
      reader.onload = function (e) {
        console.log("Изображение загружено");
        originalImage = e.target.result;
        processImage();
        downloadBtn.disabled = false; 
      };
      reader.readAsDataURL(file);
    } else {
      alert("Пожалуйста, выберите допустимый файл изображения.");
      console.log("Выбранный файл не является изображением");
    }
  });

  blocksize.addEventListener("input", () => {
    blockvalue.textContent = blocksize.value;
    console.log("Изменен размер блока:", blocksize.value);
    processImage();
  });

  useAllColors.addEventListener("input", () => {
    useAllColorsValue.textContent = useAllColors.value;
    console.log(
      "Изменена степень использования всех цветов палитры:",
      useAllColors.value
    );
    processImage();
  });

  brightness.addEventListener("input", () => {
    brightnessValue.textContent = brightness.value;
    console.log("Изменена яркость:", brightness.value);
    processImage();
  });

  contrast.addEventListener("input", () => {
    contrastValue.textContent = contrast.value;
    console.log("Изменен контраст:", contrast.value);
    processImage();
  });

  saturation.addEventListener("input", () => {
    saturationValue.textContent = saturation.value;
    console.log("Изменена насыщенность:", saturation.value);
    processImage();
  });

  contour.addEventListener("change", () => {
    console.log("Contour:", contour.checked);
    processImage();
  });

  paletteselector.addEventListener("change", () => {
    console.log("Выбрана палитра:", paletteselector.value);
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
        `Максимальное количество цветов для useAllColors: ${paletteCount}`
      );
    } else {
      useAllColors.max = 20;
      useAllColors.value = 1;
      useAllColorsValue.textContent = "1";
    }
    processImage();
  });

  // <-- Новый код: обработчик для выбора разрешения
  resolutionSelect.addEventListener("change", () => {
    console.log("Выбрано разрешение:", resolutionSelect.value);
    processImage();
  });
  // Конец нового кода

  // Функция обработки
  function processImage() {
    if (!originalImage) {
      console.log("Нет исходного изображения для обработки");
      return;
    }
    console.log("Начинаю обработку изображения");

    const img = new Image();
    img.crossOrigin = "Anonymous";
    img.onload = function () {
      const tempCanvas = document.createElement("canvas");
      tempCanvas.width = img.width;
      tempCanvas.height = img.height;
      const ctx = tempCanvas.getContext("2d");
      ctx.drawImage(img, 0, 0);
      // Отправляем текущее изображение на сервер для обработки
      uploadToServer(tempCanvas);
    };
    img.onerror = function () {
      console.error("Не удалось загрузить исходное изображение.");
    };
    img.src = originalImage;
  }

  function uploadToServer(canvas) {
    canvas.toBlob(function (blob) {
      if (!blob) {
        console.error("Не удалось создать Blob из изображения.");
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

      // <-- Новый код
      formData.append("resolution", resolutionSelect.value);
      // Конец нового кода

      console.log("Отправка данных на сервер");

      fetch("/process", {
        method: "POST",
        body: formData,
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error("Ошибка при обработке изображения.");
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
            console.log("Обработанное изображение отрисовано на canvas");
          };
          processedImg.src = url;
        })
        .catch((error) => {
          console.error("Ошибка при обработке изображения:", error);
        });
    }, "image/png");
  }

  function capitalizeFirstLetter(string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
  }
});