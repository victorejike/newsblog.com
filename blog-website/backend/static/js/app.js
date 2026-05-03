// API Base URL
const API_BASE = '/api';

// Token management
const getToken = () => localStorage.getItem('token');
const setToken = (token) => localStorage.setItem('token', token);
const removeToken = () => localStorage.removeItem('token');
const getUser = () => JSON.parse(localStorage.getItem('user') || 'null');
const setUser = (user) => localStorage.setItem('user', JSON.stringify(user));
const removeUser = () => localStorage.removeItem('user');

// Theme management
const getTheme = () => localStorage.getItem('theme') || 'light';
const setTheme = (theme) => {
    localStorage.setItem('theme', theme);
    document.documentElement.setAttribute('data-theme', theme);
    const toggle = document.getElementById('theme-toggle');
    if (toggle) {
        toggle.textContent = theme === 'dark' ? '☀️' : '🌙';
    }
};

// API helper with error handling
const apiCall = async (endpoint, options = {}) => {
    const token = getToken();
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            headers,
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'An error occurred' }));
            throw new Error(error.message || 'Request failed');
        }

        return response.json();
    } catch (error) {
        console.error('API Error:', error);
        throw error;
    }
};

// Update UI based on authentication status
const updateAuthUI = () => {
    const token = getToken();
    const user = getUser();

    const loginLink = document.getElementById('login-link');
    const registerLink = document.getElementById('register-link');
    const adminLink = document.getElementById('admin-link');
    const logoutLink = document.getElementById('logout-link');
    const bookmarksLink = document.getElementById('bookmarks-link');
    const profileLink = document.getElementById('profile-link');

    if (token && user) {
        if (loginLink) loginLink.style.display = 'none';
        if (registerLink) registerLink.style.display = 'none';
        if (adminLink && user.is_admin) adminLink.style.display = 'block';
        if (logoutLink) logoutLink.style.display = 'block';
        if (bookmarksLink) bookmarksLink.style.display = 'block';
        if (profileLink) profileLink.style.display = 'block';
    } else {
        if (loginLink) loginLink.style.display = 'block';
        if (registerLink) registerLink.style.display = 'block';
        if (adminLink) adminLink.style.display = 'none';
        if (logoutLink) logoutLink.style.display = 'none';
        if (bookmarksLink) bookmarksLink.style.display = 'none';
        if (profileLink) profileLink.style.display = 'none';
    }
};

// Load categories
const loadCategories = async () => {
    try {
        const categories = await apiCall('/categories');
        const categoriesNav = document.getElementById('categories-nav');

        if (categoriesNav && categories.length > 0) {
            categoriesNav.innerHTML = '<button class="category-chip active" data-category="all">All</button>' +
                categories.map(cat => `
                    <button class="category-chip" data-category="${cat.slug}">${escapeHtml(cat.name)}</button>
                `).join('');

            // Add click handlers
            categoriesNav.querySelectorAll('.category-chip').forEach(chip => {
                chip.addEventListener('click', () => {
                    categoriesNav.querySelectorAll('.category-chip').forEach(c => c.classList.remove('active'));
                    chip.classList.add('active');
                    const category = chip.dataset.category;
                    if (category === 'all') {
                        loadPosts();
                    } else {
                        loadCategoryPosts(category);
                    }
                });
            });
        }
    } catch (error) {
        console.error('Error loading categories:', error);
    }
};

// Load posts on home page
const loadPosts = async () => {
    try {
        const posts = await apiCall('/posts');
        const postsList = document.getElementById('posts-list');

        if (postsList) {
            if (posts.length === 0) {
                postsList.innerHTML = '<p class="text-center text-muted">No posts yet. Check back soon!</p>';
                return;
            }

            postsList.innerHTML = posts.map(post => createPostCard(post)).join('');
        }
    } catch (error) {
        console.error('Error loading posts:', error);
        const postsList = document.getElementById('posts-list');
        if (postsList) {
            postsList.innerHTML = '<p class="text-center text-muted">Error loading posts. Please try again later.</p>';
        }
    }
};

// Load category posts
const loadCategoryPosts = async (categorySlug) => {
    try {
        const posts = await apiCall(`/categories/${categorySlug}/posts`);
        const postsList = document.getElementById('posts-list');

        if (postsList) {
            if (posts.length === 0) {
                postsList.innerHTML = '<p class="text-center text-muted">No posts in this category yet.</p>';
                return;
            }

            postsList.innerHTML = posts.map(post => createPostCard(post)).join('');
        }
    } catch (error) {
        console.error('Error loading category posts:', error);
    }
};

// Load popular posts
const loadPopularPosts = async () => {
    try {
        const posts = await apiCall('/popular');
        const popularPostsList = document.getElementById('popular-posts');

        if (popularPostsList) {
            if (posts.length === 0) {
                popularPostsList.innerHTML = '<p class="text-center text-muted">No popular posts yet.</p>';
                return;
            }

            popularPostsList.innerHTML = posts.slice(0, 3).map(post => createPostCard(post)).join('');
        }
    } catch (error) {
        console.error('Error loading popular posts:', error);
    }
};

// Create post card HTML
const createPostCard = (post) => {
    const readingTime = post.reading_time ? `${post.reading_time} min read` : '';
    const coverImage = post.cover_image_url || '';
    const category = post.category || 'Uncategorized';

    return `
        <article class="post-card fade-in" data-post-id="${post.id}">
            ${coverImage ? `<img src="${escapeHtml(coverImage)}" alt="${escapeHtml(post.title)}" class="post-cover">` : ''}
            <div class="post-content">
                <span class="post-category">${escapeHtml(category)}</span>
                <h3 class="post-title">${escapeHtml(post.title)}</h3>
                <div class="post-meta">
                    <span>By ${escapeHtml(post.author)}</span>
                    <span>•</span>
                    <span>${formatDate(post.created_at)}</span>
                    ${readingTime ? `<span>•</span><span>${readingTime}</span>` : ''}
                </div>
                <p class="post-excerpt">${escapeHtml(post.excerpt || post.content)}</p>
                <div class="post-stats">
                    <span class="post-stat">👁️ ${post.view_count || 0}</span>
                    <span class="post-stat">❤️ ${post.like_count || 0}</span>
                    <span class="post-stat">💬 ${post.comment_count || 0}</span>
                </div>
            </div>
        </article>
    `;
};

// Load single post
const loadPost = async (postId) => {
    try {
        const post = await apiCall(`/posts/${postId}`);
        const postContainer = document.getElementById('post-container');

        if (postContainer) {
            const readingTime = post.reading_time ? `${post.reading_time} min read` : '';
            const category = post.category || 'Uncategorized';

            postContainer.innerHTML = `
                <div class="post-full-header">
                    ${post.cover_image_url ? `<img src="${escapeHtml(post.cover_image_url)}" alt="${escapeHtml(post.title)}" style="width: 100%; height: 400px; object-fit: cover; border-radius: 16px; margin-bottom: 24px;">` : ''}
                    <span class="post-category">${escapeHtml(category)}</span>
                    <h1 class="post-full-title">${escapeHtml(post.title)}</h1>
                    <div class="post-full-meta">
                        <span>By ${escapeHtml(post.author)}</span>
                        <span>•</span>
                        <span>${formatDate(post.created_at)}</span>
                        ${readingTime ? `<span>•</span><span>${readingTime}</span>` : ''}
                        <span>•</span>
                        <span>👁️ ${post.view_count || 0} views</span>
                    </div>
                </div>
                <div class="post-full-content">
                    ${formatContent(post.content)}
                </div>
                ${post.tags && post.tags.length > 0 ? `
                    <div class="tags-nav mt-2">
                        ${post.tags.map(tag => `<span class="tag-chip">${escapeHtml(tag)}</span>`).join('')}
                    </div>
                ` : ''}
            `;

            // Show engagement bar
            const engagementBar = document.getElementById('engagement-bar');
            if (engagementBar) {
                engagementBar.style.display = 'flex';
                document.getElementById('like-count').textContent = post.like_count || 0;
            }

            // Load comments
            loadPostComments(postId);

            // Load related posts
            loadRelatedPosts(postId);
        }
    } catch (error) {
        console.error('Error loading post:', error);
        alert('Error loading post. Please try again later.');
    }
};

// Format content with basic HTML
const formatContent = (content) => {
    // Convert line breaks to paragraphs
    return content.split('\n\n').map(para => {
        if (para.trim()) {
            return `<p>${escapeHtml(para)}</p>`;
        }
        return '';
    }).join('');
};

// Load post comments
const loadPostComments = async (postId) => {
    try {
        const comments = await apiCall(`/posts/${postId}/comments`);
        const commentsSection = document.getElementById('comments-section');
        const commentsList = document.getElementById('comments-list');
        const commentCount = document.getElementById('comment-count');

        if (commentsSection) {
            commentsSection.style.display = 'block';
        }

        if (commentCount) {
            commentCount.textContent = comments.length;
        }

        if (commentsList) {
            if (comments.length === 0) {
                commentsList.innerHTML = '<p class="text-center text-muted">No comments yet. Be the first to share your thoughts!</p>';
                return;
            }

            commentsList.innerHTML = comments.map(comment => createCommentHTML(comment)).join('');
        }

        // Show comment form if logged in
        const commentFormContainer = document.getElementById('comment-form-container');
        if (commentFormContainer && getToken()) {
            commentFormContainer.style.display = 'block';
        }
    } catch (error) {
        console.error('Error loading comments:', error);
    }
};

// Create comment HTML
const createCommentHTML = (comment) => {
    const replies = comment.replies && comment.replies.length > 0
        ? comment.replies.map(reply => createCommentHTML(reply)).join('')
        : '';

    return `
        <div class="comment fade-in">
            <div class="comment-header">
                <div class="comment-avatar">${escapeHtml(comment.username.charAt(0).toUpperCase())}</div>
                <div>
                    <div class="comment-author">${escapeHtml(comment.username)}</div>
                    <div class="comment-date">${formatDate(comment.created_at)}</div>
                </div>
            </div>
            <div class="comment-content">${escapeHtml(comment.content)}</div>
            ${replies ? `<div class="comment-replies">${replies}</div>` : ''}
        </div>
    `;
};

// Load related posts
const loadRelatedPosts = async (postId) => {
    try {
        const posts = await apiCall(`/posts/${postId}/related`);
        const relatedPostsList = document.getElementById('related-posts');

        if (relatedPostsList) {
            if (posts.length === 0) {
                relatedPostsList.innerHTML = '<p class="text-center text-muted">No related posts found.</p>';
                return;
            }

            relatedPostsList.innerHTML = posts.map(post => createPostCard(post)).join('');
        }
    } catch (error) {
        console.error('Error loading related posts:', error);
    }
};

// Load user profile
const loadUserProfile = async () => {
    try {
        const user = await apiCall('/me');
        const profileHeader = document.getElementById('profile-header');

        if (profileHeader) {
            profileHeader.innerHTML = `
                <div class="profile-avatar">${escapeHtml(user.username.charAt(0).toUpperCase())}</div>
                <div class="profile-info">
                    <h2>${escapeHtml(user.username)}</h2>
                    <p class="profile-bio">${escapeHtml(user.bio || 'No bio yet')}</p>
                    <div class="profile-links">
                        ${user.website_url ? `<a href="${escapeHtml(user.website_url)}" target="_blank" class="profile-link">🌐 Website</a>` : ''}
                        ${user.twitter_handle ? `<a href="https://twitter.com/${escapeHtml(user.twitter_handle)}" target="_blank" class="profile-link">🐦 Twitter</a>` : ''}
                    </div>
                    <button class="btn mt-1" id="edit-profile-btn">Edit Profile</button>
                </div>
            `;

            // Add edit profile handler
            document.getElementById('edit-profile-btn').addEventListener('click', () => {
                document.getElementById('profile-edit-form').style.display = 'block';
                document.getElementById('bio').value = user.bio || '';
                document.getElementById('avatar_url').value = user.avatar_url || '';
                document.getElementById('website_url').value = user.website_url || '';
                document.getElementById('twitter_handle').value = user.twitter_handle || '';
            });
        }

        // Load user's posts
        loadUserPosts(user.id);
    } catch (error) {
        console.error('Error loading user profile:', error);
    }
};

// Load user posts
const loadUserPosts = async (userId) => {
    try {
        const posts = await apiCall(`/users/${userId}/posts`);
        const userPostsList = document.getElementById('user-posts');

        if (userPostsList) {
            if (posts.length === 0) {
                userPostsList.innerHTML = '<p class="text-center text-muted">No posts yet.</p>';
                return;
            }

            userPostsList.innerHTML = posts.map(post => createPostCard(post)).join('');
        }
    } catch (error) {
        console.error('Error loading user posts:', error);
    }
};

// Load bookmarks
const loadBookmarks = async () => {
    try {
        const posts = await apiCall('/bookmarks');
        const bookmarksList = document.getElementById('bookmarks-list');
        const noBookmarks = document.getElementById('no-bookmarks');

        if (bookmarksList) {
            if (posts.length === 0) {
                bookmarksList.style.display = 'none';
                if (noBookmarks) noBookmarks.style.display = 'block';
                return;
            }

            bookmarksList.style.display = 'grid';
            if (noBookmarks) noBookmarks.style.display = 'none';
            bookmarksList.innerHTML = posts.map(post => createPostCard(post)).join('');
        }
    } catch (error) {
        console.error('Error loading bookmarks:', error);
    }
};

// Search functionality
const setupSearch = () => {
    const searchInput = document.getElementById('search-input');
    if (!searchInput) return;

    let searchTimeout;
    searchInput.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        const query = e.target.value.trim();

        if (query.length < 2) {
            loadPosts();
            return;
        }

        searchTimeout = setTimeout(async () => {
            try {
                const posts = await apiCall(`/search?q=${encodeURIComponent(query)}`);
                const postsList = document.getElementById('posts-list');

                if (postsList) {
                    if (posts.length === 0) {
                        postsList.innerHTML = '<p class="text-center text-muted">No results found for your search.</p>';
                        return;
                    }

                    postsList.innerHTML = posts.map(post => createPostCard(post)).join('');
                }
            } catch (error) {
                console.error('Error searching posts:', error);
            }
        }, 300);
    });
};

// Engagement functions
const setupEngagement = () => {
    const likeBtn = document.getElementById('like-btn');
    const bookmarkBtn = document.getElementById('bookmark-btn');
    const shareBtn = document.getElementById('share-btn');

    if (likeBtn) {
        likeBtn.addEventListener('click', async () => {
            const postId = getPostIdFromURL();
            if (!postId) return;

            try {
                if (likeBtn.classList.contains('liked')) {
                    await apiCall(`/posts/${postId}/like`, { method: 'DELETE' });
                    likeBtn.classList.remove('liked');
                    const count = parseInt(document.getElementById('like-count').textContent) - 1;
                    document.getElementById('like-count').textContent = count;
                } else {
                    await apiCall(`/posts/${postId}/like`, { method: 'POST' });
                    likeBtn.classList.add('liked');
                    const count = parseInt(document.getElementById('like-count').textContent) + 1;
                    document.getElementById('like-count').textContent = count;
                }
            } catch (error) {
                console.error('Error liking post:', error);
            }
        });
    }

    if (bookmarkBtn) {
        bookmarkBtn.addEventListener('click', async () => {
            const postId = getPostIdFromURL();
            if (!postId) return;

            try {
                if (bookmarkBtn.classList.contains('bookmarked')) {
                    await apiCall(`/posts/${postId}/bookmark`, { method: 'DELETE' });
                    bookmarkBtn.classList.remove('bookmarked');
                    bookmarkBtn.querySelector('span:last-child').textContent = 'Bookmark';
                } else {
                    await apiCall(`/posts/${postId}/bookmark`, { method: 'POST' });
                    bookmarkBtn.classList.add('bookmarked');
                    bookmarkBtn.querySelector('span:last-child').textContent = 'Bookmarked';
                }
            } catch (error) {
                console.error('Error bookmarking post:', error);
            }
        });
    }

    if (shareBtn) {
        shareBtn.addEventListener('click', () => {
            if (navigator.share) {
                navigator.share({
                    title: document.querySelector('.post-full-title')?.textContent || 'Check out this post',
                    url: window.location.href,
                });
            } else {
                // Fallback: copy to clipboard
                navigator.clipboard.writeText(window.location.href);
                alert('Link copied to clipboard!');
            }
        });
    }
};

// Comment functionality
const setupComments = () => {
    const submitCommentBtn = document.getElementById('submit-comment');
    const commentInput = document.getElementById('comment-input');

    if (submitCommentBtn && commentInput) {
        submitCommentBtn.addEventListener('click', async () => {
            const content = commentInput.value.trim();
            if (!content) {
                alert('Please enter a comment');
                return;
            }

            const postId = getPostIdFromURL();
            if (!postId) return;

            try {
                await apiCall(`/posts/${postId}/comments`, {
                    method: 'POST',
                    body: JSON.stringify({ content }),
                });

                commentInput.value = '';
                loadPostComments(postId);
            } catch (error) {
                console.error('Error posting comment:', error);
                alert('Error posting comment. Please try again.');
            }
        });
    }
};

// Get post ID from URL
const getPostIdFromURL = () => {
    const pathParts = window.location.pathname.split('/');
    const postIndex = pathParts.indexOf('post');
    if (postIndex !== -1 && postIndex + 1 < pathParts.length) {
        return parseInt(pathParts[postIndex + 1]);
    }
    return null;
};

// Register form handler
const handleRegister = async (e) => {
    e.preventDefault();

    const username = document.getElementById('username').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    try {
        const response = await apiCall('/register', {
            method: 'POST',
            body: JSON.stringify({ username, email, password }),
        });

        alert('Registration successful! Please login.');
        window.location.href = '/login';
    } catch (error) {
        console.error('Registration error:', error);
        alert(error.message || 'Registration failed. Please try again.');
    }
};

// Login form handler
const handleLogin = async (e) => {
    e.preventDefault();

    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    try {
        const response = await apiCall('/login', {
            method: 'POST',
            body: JSON.stringify({ email, password }),
        });

        setToken(response.token);
        setUser(response.user);

        if (response.user.is_admin) {
            window.location.href = '/admin';
        } else {
            window.location.href = '/';
        }
    } catch (error) {
        console.error('Login error:', error);
        alert(error.message || 'Login failed. Please check your credentials.');
    }
};

// Logout handler
const handleLogout = () => {
    removeToken();
    removeUser();
    window.location.href = '/';
};

// Profile edit handler
const handleProfileEdit = async (e) => {
    e.preventDefault();

    const bio = document.getElementById('bio').value;
    const avatar_url = document.getElementById('avatar_url').value;
    const website_url = document.getElementById('website_url').value;
    const twitter_handle = document.getElementById('twitter_handle').value;

    try {
        await apiCall('/me', {
            method: 'PUT',
            body: JSON.stringify({ bio, avatar_url, website_url, twitter_handle }),
        });

        alert('Profile updated successfully!');
        document.getElementById('profile-edit-form').style.display = 'none';
        loadUserProfile();
    } catch (error) {
        console.error('Error updating profile:', error);
        alert('Error updating profile. Please try again.');
    }
};

// Newsletter subscription
const setupNewsletter = () => {
    const newsletterForm = document.getElementById('newsletter-form');
    if (!newsletterForm) return;

    newsletterForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = newsletterForm.querySelector('input').value;

        try {
            await apiCall('/newsletter', {
                method: 'POST',
                body: JSON.stringify({ email }),
            });

            alert('Successfully subscribed to newsletter!');
            newsletterForm.reset();
        } catch (error) {
            console.error('Newsletter subscription error:', error);
            alert('Error subscribing to newsletter. Please try again.');
        }
    });
};

// Theme toggle
const setupThemeToggle = () => {
    const themeToggle = document.getElementById('theme-toggle');
    if (!themeToggle) return;

    // Initialize theme
    setTheme(getTheme());

    themeToggle.addEventListener('click', () => {
        const currentTheme = getTheme();
        const newTheme = currentTheme === 'light' ? 'dark' : 'light';
        setTheme(newTheme);
    });
};

// Utility functions
const escapeHtml = (text) => {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
};

const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
    });
};

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    // Initialize theme
    setupThemeToggle();

    // Update auth UI
    updateAuthUI();

    // Home page
    if (document.getElementById('posts-list')) {
        loadCategories();
        loadPosts();
        loadPopularPosts();
        setupSearch();
        setupNewsletter();
    }

    // Post page
    if (document.getElementById('post-container')) {
        const postId = getPostIdFromURL();
        if (postId) {
            loadPost(postId);
            setupEngagement();
            setupComments();
        }
    }

    // Profile page
    if (document.getElementById('profile-header')) {
        loadUserProfile();
    }

    // Bookmarks page
    if (document.getElementById('bookmarks-list')) {
        loadBookmarks();
    }

    // Register form
    const registerForm = document.getElementById('register-form');
    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }

    // Login form
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    // Logout link
    const logoutLink = document.getElementById('logout-link');
    if (logoutLink) {
        logoutLink.addEventListener('click', (e) => {
            e.preventDefault();
            handleLogout();
        });
    }

    // Profile edit form
    const editProfileForm = document.getElementById('edit-profile-form');
    if (editProfileForm) {
        editProfileForm.addEventListener('submit', handleProfileEdit);
    }

    // Cancel edit button
    const cancelEditBtn = document.getElementById('cancel-edit');
    if (cancelEditBtn) {
        cancelEditBtn.addEventListener('click', () => {
            document.getElementById('profile-edit-form').style.display = 'none';
        });
    }

    // Post card click handlers
    document.querySelectorAll('.post-card').forEach(card => {
        card.addEventListener('click', () => {
            const postId = card.dataset.postId;
            if (postId) {
                window.location.href = `/post/${postId}`;
            }
        });
    });
});
