const addPostButton = document.getElementById("add_post");
const post = document.getElementById("post_input");
const posts = document.getElementById("posts");

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
            likes: 0,
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
    postAuthor.innerText = post.author;

    // const postLikes = document.createElement("p");
    // postLikes.classList.add("card-text");
    // postLikes.classList.add("text-right");
    // postLikes.innerHTML = `Likes: ${post.likes}`;

    // const likeButton = document.createElement("button");
    // likeButton.classList.add("btn");
    // likeButton.classList.add("btn-primary");
    // likeButton.innerText = "Like";

    // likeButton.addEventListener("click", async () => {
        // const updatedLikes = post.likes + 1;

        // const response = await fetch(`/api/posts/${post.id}`, {
            // method: "PUT",
            // headers: {
                // "Content-Type": "application/json"
            // },
            // body: JSON.stringify({
                // likes: updatedLikes
            // }),
        // });

        // if (!response.ok) {
            // return;
        // }

        // post.likes = updatedLikes;
        // postLikes.innerHTML = `Likes: ${updatedLikes}`;
    // });

    const hr = document.createElement("hr");

    postCard.appendChild(postAuthor);
    postCard.appendChild(postContent);
    // postCard.appendChild(postLikes);
    // postCard.appendChild(likeButton);
    postBody.appendChild(postCard);
    postBody.appendChild(hr);
    posts.appendChild(postBody);
}

// const logoutLink = document.getElementById("logout_link");

// logoutLink.addEventListener("click", async () => {
//     const response = await fetch("/api/logout", {
//         method: "POST",
//         headers: {
//             "Content-Type": "application/json"
//         },
//     });
// });
