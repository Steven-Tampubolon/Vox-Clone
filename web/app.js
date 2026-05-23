const BACKEND_URL = "http://localhost:8080/api";

let mediaRecorder;
let audioChunks = [];
let audioBlob = null; // Variabel ini akan menampung blob rekaman ATAU berkas hasil upload

// Ambil element DOM
const voiceNameInput = document.getElementById('voiceName');
const btnRecord = document.getElementById('btnRecord');
const recordText = document.getElementById('recordText');
const recordIcon = document.getElementById('recordIcon');
const recordStatus = document.getElementById('recordStatus');
const audioFileInput = document.getElementById('audioFile'); // Input baru
const previewContainer = document.getElementById('previewContainer');
const audioPlayback = document.getElementById('audioPlayback');
const btnSubmitClone = document.getElementById('btnSubmitClone');
const cloneNotification = document.getElementById('cloneNotification');

const voiceSelect = document.getElementById('voiceSelect');
const btnFetchVoices = document.getElementById('btnFetchVoices');
const ttsText = document.getElementById('ttsText');
const btnGenerateTTS = document.getElementById('btnGenerateTTS');
const resultContainer = document.getElementById('resultContainer');
const ttsAudioResult = document.getElementById('ttsAudioResult');
const btnDownloadAudio = document.getElementById('btnDownloadAudio');

// --- 1. HANDLING REKAMAN SUARA MIKROFON (OPSI 1) ---
btnRecord.addEventListener('click', async () => {
    // Reset pilihan file upload jika pengguna memilih merekam langsung
    audioFileInput.value = ""; 

    if (!mediaRecorder || mediaRecorder.state === "inactive") {
        audioChunks = [];
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            mediaRecorder = new MediaRecorder(stream);
            
            mediaRecorder.ondataavailable = e => audioChunks.push(e.data);
            mediaRecorder.onstop = () => {
                audioBlob = new Blob(audioChunks, { type: 'audio/wav' });
                audioPlayback.src = URL.createObjectURL(audioBlob);
                previewContainer.classList.remove('hidden');
                
                checkFormValidity();
            };

            mediaRecorder.start();
            recordStatus.innerText = "Status: Merekam sampel suara...";
            recordText.innerText = "Hentikan Rekam";
            recordIcon.innerText = "⏹️";
            
            btnRecord.classList.remove('btn-emerald');
            btnRecord.classList.add('btn-rose');
        } catch (err) {
            alert("Akses mikrofon ditolak atau tidak didukung browser ini!");
        }
    } else {
        mediaRecorder.stop();
        mediaRecorder.stream.getTracks().forEach(track => track.stop());
        recordStatus.innerText = "Status: Rekaman selesai dikunci.";
        recordText.innerText = "Mulai Rekam";
        recordIcon.innerText = "🔴";
        
        btnRecord.classList.remove('btn-rose');
        btnRecord.classList.add('btn-emerald');
    }
});

// --- 2. HANDLING FILE UPLOAD + VALIDASI FORMAT .WAV (OPSI 2) ---
audioFileInput.addEventListener('change', (e) => {
    const file = e.target.files[0];
    
    if (!file) return;

    // VALIDASI EKSTENSI FILE & MIME TYPE
    const fileName = file.name.toLowerCase();
    const isWavFile = fileName.endsWith('.wav');

    if (!isWavFile) {
        // Tampilkan pesan error jika file bukan .wav
        alert("❌ Format file salah! Model XTTS v2 hanya menerima berkas biner berformat .wav. Silakan pilih berkas yang sesuai.");
        
        // Reset state
        audioFileInput.value = ""; // Bersihkan kolom input file
        audioBlob = null;
        previewContainer.classList.add('hidden');
        checkFormValidity();
        return;
    }

    // Jika valid, masukkan file ke dalam objek audioBlob untuk siap dikirim
    audioBlob = file; 
    
    // Tampilkan pratinjau audio agar user bisa memutar file yang di-upload
    audioPlayback.src = URL.createObjectURL(file);
    previewContainer.classList.remove('hidden');
    recordStatus.innerText = "Status: Berkas .wav berhasil dimuat.";
    
    checkFormValidity();
});

// Listener untuk mendeteksi pengetikan nama pemilik suara
voiceNameInput.addEventListener('input', () => {
    checkFormValidity();
});

// Fungsi pemusat untuk memeriksa kelayakan tombol proses
function checkFormValidity() {
    const isNameFilled = voiceNameInput.value.trim() !== "";
    const hasAudio = audioBlob !== null;

    if (isNameFilled && hasAudio) {
        btnSubmitClone.disabled = false;
        btnSubmitClone.className = "w-full py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl cursor-pointer shadow-md active:scale-95 transition";
    } else {
        btnSubmitClone.disabled = true;
        btnSubmitClone.className = "w-full py-3 bg-slate-700 text-slate-400 font-bold rounded-xl cursor-not-allowed shadow-md transition";
    }
}

// --- 3. KIRIM AUDIO KE BACKEND GOLANG (CLONE) ---
btnSubmitClone.addEventListener('click', async () => {
    const name = voiceNameInput.value.trim();
    if (!name || !audioBlob) return;

    btnSubmitClone.innerText = "Memproses di Server AI...";
    btnSubmitClone.disabled = true;

    const formData = new FormData();
    formData.append("name", name);
    // Jika dari file input, audioBlob sudah membawa nama asli, jika dari recorder kita beri nama default
    const filename = audioBlob.name || "live_recorded_sample.wav";
    formData.append("audio", audioBlob, filename);

    try {
        const response = await fetch(`${BACKEND_URL}/voices/clone`, {
            method: "POST",
            body: formData
        });

        if (!response.ok) throw new Error("Gagal memproses kloning suara di server backend Go.");

        cloneNotification.classList.remove('hidden');
        fetchVoices(); // Reload dropdown daftar suara otomatis
        
        // Reset Formulir secara keseluruhan
        voiceNameInput.value = "";
        audioFileInput.value = "";
        audioBlob = null;
        previewContainer.classList.add('hidden');
        checkFormValidity();
        recordStatus.innerText = "Status: Mikrofon Siap";
    } catch (err) {
        alert(err.message);
    } {
        btnSubmitClone.innerText = "Proses Kloning Suara";
    }
});

// --- 4. AMBIL DAFTAR SUARA YANG ADA DI MYSQL ---
async function fetchVoices() {
    try {
        const response = await fetch(`${BACKEND_URL}/voices`);
        const json = await response.json();
        
        voiceSelect.innerHTML = '<option value="">-- Pilih Suara Kloning --</option>';
        if (json.data) {
            json.data.forEach(v => {
                const option = document.createElement('option');
                option.value = v.voice_id;
                option.innerText = v.name;
                voiceSelect.appendChild(option);
            });
        }
    } catch (err) {
        console.error("Gagal mengambil data suara dari database MySQL:", err);
    }
}

btnFetchVoices.addEventListener('click', (e) => {
    e.preventDefault();
    fetchVoices();
});

// --- 5. EKSEKUSI TEXT TO SPEECH (PROSES RE-SYNTHESIS) ---
btnGenerateTTS.addEventListener('click', async () => {
    const voiceId = voiceSelect.value;
    const text = ttsText.value.trim();

    if (!voiceId || !text) {
        alert("Pilih target suara dan ketik kalimatnya terlebih dahulu!");
        return;
    }

    btnGenerateTTS.innerText = "Meresonansi Suara di Google Colab...";
    btnGenerateTTS.disabled = true;
    resultContainer.classList.add('hidden');

    try {
        const response = await fetch(`${BACKEND_URL}/voices/tts`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ voice_id: voiceId, text: text })
        });

        if (!response.ok) throw new Error("Gagal memproses pembuatan audio dari teks.");

        const audioBlobResult = await response.blob();
        const audioUrl = URL.createObjectURL(audioBlobResult);
        
        ttsAudioResult.src = audioUrl;
        btnDownloadAudio.href = audioUrl;
        resultContainer.classList.remove('hidden');
    } catch (err) {
        alert(err.message);
    } finally {
        btnGenerateTTS.innerText = "Konversi Teks ke Suara (TTS)";
        btnGenerateTTS.disabled = false;
    }
});

// Ambil data pertama kali saat dashboard dimuat
window.addEventListener('DOMContentLoaded', fetchVoices);