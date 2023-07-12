const addPostButton = document.getElementById("add_post");
const post = document.getElementById("post_input");
const posts = document.getElementById("posts");
const profileDetails = document.getElementById("profile_details");

post.addEventListener("input", () => {
    if (post.value.length > 0) {
        addPostButton.disabled = false;
    } else {
        addPostButton.disabled = true;
    }
});

addPostButton.addEventListener("click", async () => {
    const response = await fetch("/api/add-post", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            author: "",
            content: post.value,
            authorId: "",
        }),
    });

    post.value = "";
    addPostButton.disabled = true;

    const data = await response.json();

    if (!response.ok) {
        return;
    }

    createPostElement(data);
});

window.onload = async () => {
    const profileDetailsResponse = await fetch("/api/profile-details", {
        method: "GET",
        headers: {
            "Content-Type": "application/json"
        },
    });

    const profileDetailsData = await profileDetailsResponse.json();

    if (!profileDetailsResponse.ok) {
        return;
    }


    profileDetails.innerHTML += `
        <div class="profile_details">
            <div class="profile_details_name">
                <a href="/profiles/${profileDetailsData.id}" style="text-decoration: none; color: black">${profileDetailsData.name}</a>
            </div>
        </div>
    `;

    const response = await fetch("/api/posts", {
        method: "GET",
        headers: {
            "Content-Type": "application/json"
        },
    });

    const data = await response.json();

    if (!response.ok) {
        return;
    }

    data.forEach((post) => {
        createPostElement(post);
    });
}

function createPostElement(post) {
    const postBody = document.createElement("div");
    postBody.classList.add("post_body");

    const postCard = document.createElement("div");
    postCard.classList.add("card-body");

    const postContent = document.createElement("p");
    postContent.classList.add("card-text");
    postContent.classList.add("text-justify");
    postContent.innerText = post.content;

    const postAuthor = document.createElement("h5");
    postAuthor.classList.add("card-text");
    postAuthor.classList.add("text-right");
    postAuthor.innerHTML = `<a href="/profiles/${post.authorId}" style="text-decoration: none; color: black">${post.author}</a>`;

    const hr = document.createElement("hr");

    postCard.appendChild(postAuthor);
    postCard.appendChild(postContent);
    postBody.appendChild(postCard);
    postBody.appendChild(hr);
    posts.appendChild(postBody);
}
