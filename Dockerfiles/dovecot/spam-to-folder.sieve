require ["fileinto"];

if header :is "X-Spam" "Yes" {
        fileinto "Junk";
}
