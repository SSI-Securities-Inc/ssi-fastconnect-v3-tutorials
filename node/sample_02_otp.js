/**
 * Sample 02 — Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
 * ========================================================================
 * Hướng dẫn luồng Request OTP và Polling Smart OTP Push Notification.
 */

const { Auth, Config } = require('@ssi.developer/ssi-sdk');
const appConfig = require('./config');
const readline = require('readline');

const config = new Config({
  clientId: appConfig.CLIENT_ID,
  apiKey: appConfig.API_KEY,
  apiSecret: appConfig.API_SECRET,
  privateKey: appConfig.PRIVATE_KEY,
});

async function main() {
  const auth = new Auth(config);

  console.log('=== Bước 1: Yêu cầu OTP (Request OTP) ===');
  const otpRes = await auth.requestOtp();
  console.log('Request OTP Response:', otpRes);

  let transactionId = otpRes?.data?.transactionId || otpRes?.transactionId;

  if (transactionId) {
    console.log(`\n[Smart OTP] Đã nhận Transaction ID: ${transactionId}`);
    console.log('Vui lòng mở ứng dụng SSI trên điện thoại và bấm APPROVE (Xác nhận)...');
    console.log('SDK đang Polling chờ bạn bấm phê duyệt...');

    try {
      const accessToken = await auth.ensureAuthenticated(undefined, transactionId, 5000, 6);
      console.log('\n[THÀNH CÔNG] Đã xác thực Smart OTP!');
      console.log('Access Token:', accessToken.slice(0, 40), '...');
    } catch (err) {
      console.error('\n[LỖI/TIMEOUT] Phê duyệt Smart OTP thất bại:', err.message);
    }
  } else {
    console.log('\n[OTP Thường / Smart OTP lấy trực tiếp trên App]');
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
    rl.question('Vui lòng nhập mã OTP 6 số: ', async (userOtp) => {
      rl.close();
      if (userOtp.trim()) {
        const token = await auth.authenticate(userOtp.trim());
        console.log('\n[THÀNH CÔNG] Đã xác thực mã OTP!');
        console.log('Access Token:', token.accessToken.slice(0, 40), '...');
      }
    });
  }
}

main().catch(console.error);
